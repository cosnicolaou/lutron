package schedule

/*
import (
	"context"
	"fmt"
	"os"

	"cloudeng.io/cmdutil/cmdyaml"
)

// AuthConfig represents a specific auth configuration and is intended
// to be reused and referred to by it's auth_id.
type AuthConfig struct {
	ID    string `yaml:"auth_id"`
	User  string `yaml:"user"`
	Token string `yaml:"token"`
}

// AuthInfo is a map of ID/auth_id to AuthConfig.
type AuthInfo map[string]AuthConfig

func (a AuthConfig) String() string {
	return a.ID + "[" + a.User + "]	"
}

// ParseAuthConfigFile parses an auth file into an AuthInfo map and stores that
// in the returned context.
func ParseAuthConfigFile(ctx context.Context, filename string) (context.Context, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return ctx, err
	}
	return ParseAuthConfig(ctx, data)
}

func ParseAuthConfig(ctx context.Context, data []byte) (context.Context, error) {
	var auth []AuthConfig
	if err := cmdyaml.ParseConfig(data, &auth); err != nil {
		return ctx, err
	}
	am := AuthInfo{}
	for _, a := range auth {
		if a.ID == "" {
			return ctx, fmt.Errorf("auth_id is required")
		}
		if a.User == "" {
			return ctx, fmt.Errorf("user is required")
		}
		if a.Token == "" {
			return ctx, fmt.Errorf("token is required")
		}
		am[a.ID] = a
	}
	return ContextWithAuth(ctx, am), nil
}

type authKey struct{}

func ContextWithAuth(ctx context.Context, am AuthInfo) context.Context {
	return context.WithValue(ctx, authKey{}, am)
}

func AuthFromContextForID(ctx context.Context, id string) AuthConfig {
	am, ok := ctx.Value(authKey{}).(AuthInfo)
	if !ok {
		return AuthConfig{}
	}
	return am[id]
}
*/
