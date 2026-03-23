package auth

type Encoder interface {
	Encrypt(salt []byte, pwd string) (string, error)
	Matches(pwd string, hashedPwd string) bool
}

type Argon2PasswordEncoder struct {
	argon2Hasher *Argon2Hasher
}

func (a Argon2PasswordEncoder) Encrypt(salt []byte, pwd string) (string, error) {
	return a.argon2Hasher.Hash(pwd, salt)
}

func (a Argon2PasswordEncoder) Matches(pwd string, hashedPwd string) bool {
	return a.argon2Hasher.CompareHashAndPassword(hashedPwd, pwd) == nil
}

var pwdEncoder Encoder = &Argon2PasswordEncoder{
	argon2Hasher: NewArgon2Hasher(19, 64*1024, 3, 4, 32),
}
