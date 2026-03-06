/**
 * Minimal YAML parser for role.yaml files.
 * Handles: top-level scalars, top-level lists, one-level nested objects with lists.
 */
export function parse(text: string): Record<string, unknown> {
  const result: Record<string, unknown> = {}
  const lines = text.split('\n')

  let topKey = ''
  let topList: string[] | null = null
  let nestedObj: Record<string, unknown> | null = null
  let nestedList: string[] | null = null

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i].replace(/\r$/, '')
    if (/^\s*#/.test(line) || /^\s*$/.test(line)) continue

    const indent = line.search(/\S/)

    // Indent 4+: list item under nested key
    if (indent >= 4) {
      const m = line.match(/^\s+-\s+"?(.+?)"?\s*$/)
      if (m && nestedList) nestedList.push(m[1])
      continue
    }

    // Indent 2-3: nested key or list item under top key
    if (indent >= 2) {
      const listItem = line.match(/^\s+-\s+"?(.+?)"?\s*$/)
      if (listItem && topList) {
        topList.push(listItem[1])
        continue
      }

      const keyMatch = line.match(/^\s+(\w[\w_]*):\s*(.*)$/)
      if (keyMatch && nestedObj) {
        const val = keyMatch[2].trim()
        if (!val) {
          const list: string[] = []
          nestedObj[keyMatch[1]] = list
          nestedList = list
          topList = null
        } else {
          nestedObj[keyMatch[1]] = unquote(val)
          nestedList = null
        }
      }
      continue
    }

    // Indent 0: top-level key
    const topMatch = line.match(/^(\w[\w_]*):\s*(.*)$/)
    if (!topMatch) continue

    topKey = topMatch[1]
    const val = topMatch[2].trim()
    nestedObj = null
    nestedList = null
    topList = null

    if (!val) {
      // Peek at next non-empty line to decide list vs object
      let isObject = false
      for (let j = i + 1; j < lines.length; j++) {
        const next = lines[j].replace(/\r$/, '')
        if (/^\s*#/.test(next) || /^\s*$/.test(next)) continue
        isObject = /^\s+\w+[\w_]*:/.test(next)
        break
      }
      if (isObject) {
        const obj: Record<string, unknown> = {}
        result[topKey] = obj
        nestedObj = obj
      } else {
        const list: string[] = []
        result[topKey] = list
        topList = list
      }
    } else {
      result[topKey] = unquote(val)
    }
  }

  return result
}

function unquote(s: string): string {
  if ((s.startsWith('"') && s.endsWith('"')) || (s.startsWith("'") && s.endsWith("'"))) {
    return s.slice(1, -1)
  }
  return s
}
