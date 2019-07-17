package foon

type FoonError string

const (
	NoSuchDocument FoonError = "NoSuchEntity"
	InvalidId      FoonError = "InvalidID"
)

func (f FoonError) Error() string {
	return string(f)
}

func (f FoonError) Is(err error) bool {
	if err == nil {
		return false
	}
	return err.Error() == f.Error()
}

func (f FoonError) IsNot(err error) bool {
	return !f.Is(err)
}
