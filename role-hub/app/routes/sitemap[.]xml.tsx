import { getRoles, getRepos } from '~/server/roles.server'
import { buildCanonical, getSiteUrl } from '~/utils/seo'

interface UrlEntry {
  loc: string
  changefreq: string
  priority: string
}

function renderUrl({ loc, changefreq, priority }: UrlEntry): string {
  return `  <url>
    <loc>${loc}</loc>
    <changefreq>${changefreq}</changefreq>
    <priority>${priority}</priority>
  </url>`
}

export async function loader() {
  const [rolesResult, repos] = await Promise.all([getRoles({ limit: 1000 }), getRepos()])

  const urls: UrlEntry[] = [
    { loc: buildCanonical('/'), changefreq: 'daily', priority: '1.0' },
    ...rolesResult.data
      .filter((r) => r.status === 'verified')
      .map((r) => ({
        loc: `${getSiteUrl()}/roles/${encodeURIComponent(r.role_name)}`,
        changefreq: 'weekly',
        priority: '0.8',
      })),
    ...repos.map((r) => ({
      loc: `${getSiteUrl()}/repos/${encodeURIComponent(r.owner)}/${encodeURIComponent(r.repo)}`,
      changefreq: 'weekly',
      priority: '0.7',
    })),
  ]

  const xml = `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
${urls.map(renderUrl).join('\n')}
</urlset>`

  return new Response(xml, {
    headers: {
      'Content-Type': 'application/xml; charset=utf-8',
      'Cache-Control': 'public, max-age=3600',
    },
  })
}
