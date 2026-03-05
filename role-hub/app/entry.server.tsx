import { PassThrough } from 'node:stream'
import type { EntryContext } from '@remix-run/node'
import { RemixServer } from '@remix-run/react'
import { renderToPipeableStream } from 'react-dom/server'

const ABORT_DELAY = 5000

export default function handleRequest(
  request: Request,
  statusCode: number,
  headers: Headers,
  context: EntryContext,
) {
  return new Promise<Response>((resolve, reject) => {
    let shellRendered = false

    const { pipe, abort } = renderToPipeableStream(
      <RemixServer context={context} url={request.url} />,
      {
        onShellReady() {
          shellRendered = true
          const body = new PassThrough()

          headers.set('Content-Type', 'text/html')

          resolve(new Response(body, { status: statusCode, headers }))

          pipe(body)
        },
        onShellError(error) {
          reject(error)
        },
        onError(error) {
          statusCode = 500
          if (shellRendered) {
            console.error(error)
          }
        },
      },
    )

    setTimeout(abort, ABORT_DELAY)
  })
}
