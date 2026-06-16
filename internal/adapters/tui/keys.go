package tui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	TabS3      key.Binding
	TabSQS     key.Binding
	TabSNS     key.Binding
	TabSecrets key.Binding
	TabConfig  key.Binding
	Up         key.Binding
	Down       key.Binding
	Left       key.Binding
	Right      key.Binding
	Enter      key.Binding
	Esc        key.Binding
	Quit       key.Binding
	CommandLog key.Binding
	Filter     key.Binding
	Sort       key.Binding

	S3Presign  key.Binding
	S3Delete   key.Binding
	S3Download key.Binding
	S3Create   key.Binding
	S3Upload   key.Binding
	S3Preview  key.Binding
	S3Version  key.Binding
	S3Folder   key.Binding

	SQSPurge    key.Binding
	SQSPurgeAll key.Binding
	SQSView     key.Binding
	SQSSend     key.Binding
	SQSDelete   key.Binding
	SQSCreate   key.Binding

	SNSCreate     key.Binding
	SNSDelete     key.Binding
	SNSEdit       key.Binding
	SNSBatch      key.Binding
	SNSImportYaml key.Binding

	SecretsCreate  key.Binding
	SecretsUpdate  key.Binding
	SecretsDelete  key.Binding
	SecretsReveal  key.Binding
	SecretsRestore key.Binding

	SQSBatchSubscribe key.Binding
	SQSSubDelete      key.Binding

	Profile       key.Binding
	ProfileExport key.Binding
	ProfileImport key.Binding
	CopyARN       key.Binding
	CommandPalette key.Binding
	ToggleTheme    key.Binding
}

var keys = keyMap{
	TabS3: key.NewBinding(
		key.WithKeys("1"),
		key.WithHelp("1", "S3 Explorer"),
	),
	TabSQS: key.NewBinding(
		key.WithKeys("2"),
		key.WithHelp("2", "SQS Queues"),
	),
	TabSNS: key.NewBinding(
		key.WithKeys("3"),
		key.WithHelp("3", "SNS Topics"),
	),
	TabSecrets: key.NewBinding(
		key.WithKeys("4"),
		key.WithHelp("4", "Secrets"),
	),
	TabConfig: key.NewBinding(
		key.WithKeys("5"),
		key.WithHelp("5", "Settings"),
	),
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "back/left"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "forward/right"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select/confirm"),
	),
	Esc: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back/close"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	CommandLog: key.NewBinding(
		key.WithKeys("o"),
		key.WithHelp("o", "command logs"),
	),
	Filter: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "filter list"),
	),
	Sort: key.NewBinding(
		key.WithKeys("ctrl+s"),
		key.WithHelp("ctrl+s", "sort"),
	),
	S3Presign: key.NewBinding(
		key.WithKeys("b"),
		key.WithHelp("b", "open in browser"),
	),
	S3Delete: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "delete"),
	),
	S3Download: key.NewBinding(
		key.WithKeys("w"),
		key.WithHelp("w", "download file"),
	),
	S3Create: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "create bucket"),
	),
	S3Upload: key.NewBinding(
		key.WithKeys("u"),
		key.WithHelp("u", "upload object"),
	),
	S3Preview: key.NewBinding(
		key.WithKeys("v"),
		key.WithHelp("v", "preview object"),
	),
	S3Version: key.NewBinding(
		key.WithKeys("V", "ctrl+v"),
		key.WithHelp("V", "object versions"),
	),
	S3Folder: key.NewBinding(
		key.WithKeys("f"),
		key.WithHelp("f", "create folder"),
	),
	SQSPurge: key.NewBinding(
		key.WithKeys("m"),
		key.WithHelp("m", "purge queue"),
	),
	SQSPurgeAll: key.NewBinding(
		key.WithKeys("M"),
		key.WithHelp("M", "purge all"),
	),
	SQSView: key.NewBinding(
		key.WithKeys("v"),
		key.WithHelp("v", "peek messages"),
	),
	SQSSend: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "publish message"),
	),
	SQSDelete: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "delete queue"),
	),
	SQSCreate: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "create queue"),
	),
	SNSCreate: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "create"),
	),
	SNSDelete: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "delete"),
	),
	SNSEdit: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "edit filter"),
	),
	SNSBatch: key.NewBinding(
		key.WithKeys("b"),
		key.WithHelp("b", "batch subscribe"),
	),
	SNSImportYaml: key.NewBinding(
		key.WithKeys("i"),
		key.WithHelp("i", "import YAML"),
	),
	SecretsCreate: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "create secret"),
	),
	SecretsUpdate: key.NewBinding(
		key.WithKeys("u"),
		key.WithHelp("u", "update secret"),
	),
	SecretsDelete: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "delete secret"),
	),
	SecretsReveal: key.NewBinding(
		key.WithKeys("v"),
		key.WithHelp("v", "reveal value"),
	),
	SecretsRestore: key.NewBinding(
		key.WithKeys("R"),
		key.WithHelp("R", "recover deleted secret"),
	),
	SQSBatchSubscribe: key.NewBinding(
		key.WithKeys("b"),
		key.WithHelp("b", "subscribe SNS topics"),
	),
	SQSSubDelete: key.NewBinding(
		key.WithKeys("d", "x"),
		key.WithHelp("d/x", "unsubscribe"),
	),
	Profile: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "profile switcher"),
	),
	ProfileExport: key.NewBinding(
		key.WithKeys("E"),
		key.WithHelp("E", "export snapshot"),
	),
	ProfileImport: key.NewBinding(
		key.WithKeys("L"),
		key.WithHelp("L", "load snapshot"),
	),
	CopyARN: key.NewBinding(
		key.WithKeys("ctrl+y"),
		key.WithHelp("ctrl+y", "copy ARN"),
	),
	CommandPalette: key.NewBinding(
		key.WithKeys("ctrl+p"),
		key.WithHelp("ctrl+p", "command palette"),
	),
	ToggleTheme: key.NewBinding(
		key.WithKeys("T"),
		key.WithHelp("T", "toggle light/dark theme"),
	),
}
