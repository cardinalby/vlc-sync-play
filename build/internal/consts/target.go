package consts

type Os string

const (
	OsWindows Os = "windows"
	OsDarwin  Os = "darwin"
	OsLinux   Os = "linux"
)

type Arch string

const (
	ArchAmd64 Arch = "amd64"
	ArchArm64 Arch = "arm64"
)
