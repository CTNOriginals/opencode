import type { Plugin } from "@opencode-ai/plugin"

export default (async ({ client, $ }) => {
	return {
		event: async ({ event }) => {
			if (event.type !== "session.idle") return

			const sessionID = event.properties.sessionID
			try {
				const result = await client.session.messages({
					path: { id: sessionID },
					query: { limit: 10 },
				})

				const messages = result.data
				if (!messages || messages.length === 0) return

				const lines: string[] = []
				for (const msg of messages) {
					const role = msg.info.role
					const textParts = msg.parts.filter(
						(p: any) => p.type === "text"
					)
					for (const part of textParts) {
						const text = (part as any).text as string
						if (text && text.trim()) {
							lines.push(`[${role}] ${text.trim()}`)
						}
					}
				}

				if (lines.length === 0) return

				const context = lines.join("\n\n")
				const proc = $`/home/ctn/.config/opencode/custom/memory/write-session.sh`
				proc.stdin.write(context)
				proc.stdin.end()
				await proc
			} catch (err) {
				// Memory is non-critical; swallow errors silently
			}
		},
	}
}) satisfies Plugin
