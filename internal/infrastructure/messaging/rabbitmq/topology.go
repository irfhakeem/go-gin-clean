package rabbitmq

var MainQueues = map[string][]string{
	"email": {
		"user.register",
		"user.reset_password",
	},
}
