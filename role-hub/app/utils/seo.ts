const DEFAULT_SITE_URL = 'https://role-hub.dev'

export function getSiteUrl(): string {
  return (typeof process !== 'undefined' && process.env.SITE_URL) || DEFAULT_SITE_URL
}

interface MetaOptions {
  title: string
  description: string
  canonical: string
  ogImage?: string
}

export function buildMeta({ title, description, canonical, ogImage }: MetaOptions) {
  const fullTitle = title === 'Role Hub' ? title : `${title} | Role Hub`
  return [
    { title: fullTitle },
    { name: 'description', content: description },
    { tagName: 'link', rel: 'canonical', href: canonical },
    { property: 'og:type', content: 'website' },
    { property: 'og:title', content: fullTitle },
    { property: 'og:description', content: description },
    { property: 'og:url', content: canonical },
    ...(ogImage ? [{ property: 'og:image', content: ogImage }] : []),
    { name: 'twitter:card', content: 'summary' },
    { name: 'twitter:title', content: fullTitle },
    { name: 'twitter:description', content: description },
  ] as const
}

export function buildCanonical(path: string): string {
  return `${getSiteUrl()}${path}`
}

export function buildJsonLd(schema: Record<string, unknown>): Record<string, unknown> {
  return schema
}
