import type { ActionFunctionArgs } from '@remix-run/node'
import { json } from '@remix-run/node'
import { handleIngestRequest } from '~/server/ingest.server'

export async function action({ request }: ActionFunctionArgs) {
  return handleIngestRequest(request)
}

export async function loader() {
  return json({ message: 'Method Not Allowed' }, { status: 405 })
}
