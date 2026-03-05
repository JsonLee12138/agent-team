import type { LoaderFunctionArgs } from '@remix-run/node'
import { json } from '@remix-run/node'
import { getRepo } from '~/server/roles.server'
import { getErrorMessage } from '~/utils/errors'

export async function loader({ params }: LoaderFunctionArgs) {
  const owner = params.owner
  const repo = params.repo
  if (!owner || !repo) {
    return json({ message: 'Repository owner and name are required.' }, { status: 400 })
  }

  try {
    const repository = await getRepo(owner, repo)
    if (!repository) {
      return json({ message: 'Repository not found.' }, { status: 404 })
    }
    return json(repository)
  } catch (error) {
    return json({ message: getErrorMessage(error) }, { status: 500 })
  }
}
