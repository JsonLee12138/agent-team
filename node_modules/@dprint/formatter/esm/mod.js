import * as v3 from "./v3.js";
import * as v4 from "./v4.js";
/**
 * Creates a formatter context for managing multiple plugins with shared configuration.
 * @param globalConfig - Global configuration shared across all plugins.
 */
export function createContext(globalConfig = {}) {
    const plugins = [];
    function findPluginForFile(filePath) {
        const fileName = getFileName(filePath);
        const ext = getFileExtension(filePath);
        // First try to match by exact file name
        for (const plugin of plugins) {
            if (plugin.fileNames.has(fileName)) {
                return plugin;
            }
        }
        // Then try to match by extension
        if (ext) {
            for (const plugin of plugins) {
                if (plugin.fileExtensions.has(ext)) {
                    return plugin;
                }
            }
        }
        return undefined;
    }
    function createHostFormatter(currentPlugin) {
        return (request) => {
            const plugin = findPluginForFile(request.filePath);
            if (plugin && plugin !== currentPlugin) {
                return plugin.formatter.formatText(request);
            }
            // Return unchanged if no other plugin matches
            return request.fileText;
        };
    }
    return {
        async addPluginStreaming(source, pluginConfig) {
            const wasmModule = await createWasmModuleFromStreaming(source);
            return this.addPlugin(wasmModule, pluginConfig);
        },
        addPlugin(source, pluginConfig = {}) {
            const formatter = source instanceof WebAssembly.Module
                ? createFromWasmModule(source)
                : createFromBuffer(source);
            // Set configuration
            formatter.setConfig(globalConfig, pluginConfig);
            // Get file matching info
            const matchingInfo = formatter.getFileMatchingInfo();
            const fileExtensions = new Set(matchingInfo.fileExtensions.map((ext) => ext.toLowerCase()));
            const fileNames = new Set(matchingInfo.fileNames.map((name) => name.toLowerCase()));
            const registered = {
                formatter,
                pluginConfig,
                fileExtensions,
                fileNames,
            };
            plugins.push(registered);
            // Set up host formatter for this plugin
            formatter.setHostFormatter(createHostFormatter(registered));
            // Return a context-aware formatter
            return {
                formatText(request) {
                    return formatter.formatText(request);
                },
                getResolvedConfig() {
                    return formatter.getResolvedConfig();
                },
                getConfigDiagnostics() {
                    return formatter.getConfigDiagnostics();
                },
            };
        },
        formatText(request) {
            const plugin = findPluginForFile(request.filePath);
            if (!plugin) {
                throw new Error(`No plugin found for file: ${request.filePath}. `
                    + `Registered plugins handle: ${plugins
                        .map((p) => [...p.fileExtensions].join(", "))
                        .join("; ")}`);
            }
            return plugin.formatter.formatText(request);
        },
        getConfigDiagnostics() {
            return plugins.flatMap((p) => p.formatter.getConfigDiagnostics());
        },
    };
}
function getFileName(filePath) {
    const lastSlash = Math.max(filePath.lastIndexOf("/"), filePath.lastIndexOf("\\"));
    return (lastSlash >= 0 ? filePath.slice(lastSlash + 1) : filePath)
        .toLowerCase();
}
function getFileExtension(filePath) {
    const fileName = getFileName(filePath);
    const lastDot = fileName.lastIndexOf(".");
    if (lastDot > 0) {
        return fileName.slice(lastDot + 1);
    }
    return undefined;
}
/**
 * Creates a formatter from the specified streaming source.
 * @remarks This is the most efficient way to create a formatter.
 * @param response - The streaming source to create the formatter from.
 */
export async function createStreaming(responsePromise) {
    const wasmModule = await createWasmModuleFromStreaming(responsePromise);
    return createFromWasmModule(wasmModule);
}
async function createWasmModuleFromStreaming(responsePromise) {
    const response = await responsePromise;
    if (response.status !== 200) {
        throw new Error(`Unexpected status code: ${response.status}\n${await response.text()}`);
    }
    if (typeof WebAssembly.instantiateStreaming === "function"
        && response.headers.get("content-type") === "application/wasm") {
        // deno-lint-ignore no-explicit-any
        return await WebAssembly.compileStreaming(response);
    }
    else {
        // fallback for node.js or when the content type isn't application/wasm
        return response.arrayBuffer()
            .then((buffer) => new WebAssembly.Module(buffer));
    }
}
/**
 * Creates a formatter from the specified wasm module bytes.
 * @param wasmModuleBuffer - The buffer of the wasm module.
 */
export function createFromBuffer(wasmModuleBuffer) {
    const wasmModule = new WebAssembly.Module(wasmModuleBuffer);
    return createFromWasmModule(wasmModule);
}
export function createFromWasmModule(wasmModule) {
    const version = getModuleVersionOrThrow(wasmModule);
    if (version === 3) {
        const host = v3.createHost();
        const wasmInstance = new WebAssembly.Instance(wasmModule, host.createImportObject());
        return v3.createFromInstance(wasmInstance, host);
    }
    else {
        const _assert4 = version;
        const host = v4.createHost();
        const wasmInstance = new WebAssembly.Instance(wasmModule, host.createImportObject());
        return v4.createFromInstance(wasmInstance, host);
    }
}
function getModuleVersionOrThrow(module) {
    const version = getModuleVersion(module);
    if (version == null) {
        throw new Error("Couldn't determine dprint plugin version. Maybe the js-formatter version is too old?");
    }
    else if (version === 3 || version === 4) {
        return version;
    }
    else if (version > 4) {
        throw new Error(`Unsupported new dprint plugin version '${version}'. Maybe the js-formatter version is too old?`);
    }
    else {
        throw new Error(`Unsupported old dprint plugin version '${version}'. Please upgrade the plugin.`);
    }
}
function getModuleVersion(module) {
    function getVersionFromExport(name) {
        if (name === "get_plugin_schema_version") {
            return 3;
        }
        const prefix = "dprint_plugin_version_";
        if (name.startsWith(prefix)) {
            const value = parseInt(name.substring(prefix.length), 10);
            if (!isNaN(value)) {
                return value;
            }
        }
        return undefined;
    }
    const exports = WebAssembly.Module.exports(module);
    for (const e of exports) {
        const maybeVersion = getVersionFromExport(e.name);
        if (maybeVersion != null) {
            return maybeVersion;
        }
    }
    return undefined;
}
