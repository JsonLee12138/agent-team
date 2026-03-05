import type { LoaderFunctionArgs } from '@remix-run/node'
import { json } from '@remix-run/node'
import { getRoleByName } from '~/server/roles.server'
import { getErrorMessage } from '~/utils/errors'

export async function loader({ params }: LoaderFunctionArgs) {
  const name = params.name
  if (!name) {
    return json({ message: 'Role name is required.' }, { status: 400 })
  }

  try {
    const role = await getRoleByName(name)
    if (!role) {
      return json({ message: 'Role not found.' }, { status: 404 })
    }
    return json(role)
  } catch (error) {
    return json({ message: getErrorMessage(error) }, { status: 500 })
  }
}
