package Alfred

type AlfredError string

func (ae AlfredError) Error() string {
    return string(ae)
}
