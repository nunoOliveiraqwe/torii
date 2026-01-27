package microProxy

// Version is a constant variable containing the version number
var Version = "0.0.x"

// Build is the unique number based off the git commit in which it is compiled against
var Build = "non commited"

var DefaultPassword = ""

var BuildTime = "unknown"

type Application struct {
	flags *Flags
}

func NewApplication() *Application {
	return &Application{
		flags: RegisterFlags(),
	}
}
