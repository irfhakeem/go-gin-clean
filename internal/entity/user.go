package entity

type Role string

const (
	RoleAdmin Role = "Admin"
	RoleUser  Role = "User"
)

type Gender string

const (
	Male   Gender = "Male"
	Female Gender = "Female"
	Other  Gender = "Other"
)

func (g Gender) IsValid() bool {
	switch g {
	case Male, Female, Other:
		return true
	default:
		return false
	}
}

func (g Gender) String() string {
	return string(g)
}

type User struct {
	PKID       int64  ` gorm:"primaryKey;autoIncrement;column:pkid"`
	Code       string ` gorm:"uniqueIndex;type:varchar(11);not null"`
	Name       string ` gorm:"not null"`
	Email      string ` gorm:"type:varchar(100);uniqueIndex;not null;column:email"`
	Password   string ` gorm:"type:varchar(255);column:password"`
	Avatar     string ` gorm:"default:''"`
	Gender     Gender ` gorm:"type:gender;default:null"`
	Role       Role   ` gorm:"type:role;default:'User'"`
	IsActive   bool   ` gorm:"default:true;not null"`
	IsVerified bool   ` gorm:"default:false;not null"`

	OAuthProvider string `gorm:"type:varchar(50);column:oauth_provider"`
	OAuthID       string `gorm:"type:varchar(255);column:oauth_id"`

	Audit
}

func (User) TableName() string {
	return "users"
}

func NewUser(name, email, password, avatar string, Gender Gender) (*User, error) {
	return &User{
		Name:     name,
		Email:    email,
		Password: password,
		Avatar:   avatar,
		Gender:   Gender,
	}, nil
}

func NewUserFromOAuth(name, emailStr, oauthProvider, oauthID, avatar string) (*User, error) {
	return &User{
		Name:          name,
		Email:         emailStr,
		Password:      "",
		Avatar:        avatar,
		OAuthProvider: oauthProvider,
		OAuthID:       oauthID,
		IsVerified:    true,
		Gender:        Other,
		IsActive:      true,
	}, nil
}

func (u *User) Equals(other *User) bool {
	if other == nil {
		return false
	}

	return u.PKID == other.PKID &&
		u.Name == other.Name &&
		u.Email == other.Email
}

func (u *User) SetEmail(email string) {
	u.Email = email
}

func (u *User) SetPassword(hashedPassword string) {
	u.Password = hashedPassword
}

func (u *User) Activate() {
	u.IsActive = true
}

func (u *User) Deactivate() {
	u.IsActive = false
}

func (u *User) VerifyEmail() {
	u.IsVerified = true
}

func (u *User) IsOAuthUser() bool {
	return u.OAuthProvider != "" && u.OAuthID != ""
}
