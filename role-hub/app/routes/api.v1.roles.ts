import type { LoaderFunctionArgs } from '@remix-run/node'
import { json } from '@remix-run/node'
import { getRoles } from '~/server/roles.server'
import { getErrorMessage } from '~/utils/errors'
import type { RolesQueryParams } from '~/types'

export async function loader({ request }: LoaderFunctionArgs) {
  const url = new URL(request.url)
  const params: RolesQueryParams = {
    search: url.searchParams.get('search') ?? undefined,
    category: url.searchParams.get('category') ?? undefined,
    framework: url.searchParams.getAll('framework'),
    sort: (url.searchParams.get('sort') as RolesQueryParams['sort']) ?? undefined,
    page: url.searchParams.get('page') ? Number(url.searchParams.get('page')) : undefined,
    limit: url.searchParams.get('limit') ? Number(url.searchParams.get('limit')) : undefined,
  }

  try {
    const data = await getRoles(params)
    return json(data)
  } catch (error) {
    return json({ message: getErrorMessage(error) }, { status: 500 })
  }
}
