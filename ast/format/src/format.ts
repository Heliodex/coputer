import { spawn, spawnSync } from "node:child_process"

export function formatFile(tool: string, path: string) {
	const proc = spawnSync(tool, ["file", path])

	if (proc.error) throw proc.error
	return proc.output[1]?.toString() || ""
}

export async function formatContent(tool: string, content: string) {
	const proc = spawn(tool, ["input"])
	proc.stdin?.write(content)
	proc.stdin?.end()

	const arr: Buffer[] = []
	proc.stdout.on("data", data => {
		arr.push(data as Buffer)
	})

	await new Promise(r => proc.on("close", r))

	if (proc.exitCode !== 0)
		throw new Error(`Formatter exited with code ${proc.exitCode}`)
	if (!proc.stdout) throw new Error("No output from formatter")

	return Buffer.concat(arr).toString()
}
