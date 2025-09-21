// The module 'vscode' contains the VS Code extensibility API
// Import the module and reference it with the alias vscode in your code below
import * as vscode from "vscode"
import { formatContent, formatFile } from "./format"

// This method is called when your extension is activated
// Your extension is activated the very first time the command is executed
async function format(
	document: vscode.TextDocument
): Promise<vscode.TextEdit[]> {
	vscode.window.showInformationMessage("Formatting document...")

	const config = vscode.workspace.getConfiguration("coputer")
	const tool = config.get<string>("tools.path")

	if (!tool) {
		vscode.window.showErrorMessage("Coputer tools path is not configured.")
		return []
	}

	// Disk files can be formatted directly, content files will need to be written to a temp file by the formatter
	const isDiskFile = document.uri.scheme === "file"
	let formatted: string
	try {
		formatted = isDiskFile
			? formatFile(tool, document.uri.fsPath)
			: await formatContent(tool, document.getText())
	} catch (err) {
		const e = err as Error
		vscode.window.showErrorMessage(
			`Failed to format document: ${e.message}`
		)
		return []
	}

	const end = document.lineAt(document.lineCount - 1)
	const replaced = vscode.TextEdit.replace(
		new vscode.Range(new vscode.Position(0, 0), end.range.end),
		formatted
	)
	return [replaced]
}

export function activate(context: vscode.ExtensionContext) {
	// Use the console to output diagnostic information (console.log) and errors (console.error)
	// This line of code will only be executed once when your extension is activated
	vscode.window.showInformationMessage("Coputer extension is now active")
	const disposable = vscode.languages.registerDocumentFormattingEditProvider(
		"luau",
		{ provideDocumentFormattingEdits: format }
	)

	context.subscriptions.push(disposable)
}

// This method is called when your extension is deactivated
export function deactivate() {}
