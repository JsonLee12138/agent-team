import { json } from '@remix-run/node'
import { getRepos } from '~/server/roles.server'
import { getErrorMessage } from '~/utils/errors'

export async function loader() {
  try {
    const repos = await getRepos()
    return json(repos)
  } catch (error) {
    return json({ message: getErrorMessage(error) }, { status: 500 })
  }
}
