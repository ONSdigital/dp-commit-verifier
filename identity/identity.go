// Package identity provides functionality for validating the identity of a
// commit author.
package identity

import (
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

var client = func() *github.Client {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: os.Getenv("GITHUB_ACCESS_TOKEN")})
	oc := oauth2.NewClient(oauth2.NoContext, ts)

	return github.NewClient(oc)
}()

type identity struct {
	Payload   string
	Signature string
	Verified  bool
}

// IsValid returns the validity of the identity of a commit author.
func IsValid(owner, name, id string) (bool, error) {
	i, err := getIdentity(owner, name, id)
	if err != nil {
		return false, err
	}
	if i == nil || i.Verified == false {
		return false, nil
	}

	v, err := verify(i.Signature, i.Payload)
	if err != nil {
		return false, err
	}
	if !v {
		return false, nil
	}

	return true, nil
}

func getIdentity(owner, respository, commit string) (*identity, error) {
	c, _, err := client.Git.GetCommit(owner, respository, commit)
	if err != nil {
		return nil, err
	}
	if c.Verification == nil || c.Verification.Payload == nil || c.Verification.Signature == nil {
		return nil, nil
	}

	return &identity{
		Payload:   *c.Verification.Payload,
		Signature: *c.Verification.Signature,
		Verified:  *c.Verification.Verified,
	}, nil
}

func verify(signature, payload string) (bool, error) {
	f, err := ioutil.TempFile(os.TempDir(), "commit-signature")
	if err != nil {
		return false, err
	}
	defer func() {
		f.Close()
		os.Remove(f.Name())
	}()

	if _, err = f.Write([]byte(signature)); err != nil {
		return false, err
	}

	cmd := exec.Command("gpg", "--status-fd=1", "--keyid-format=long", "--verify", f.Name(), "-")

	o, err := cmd.StdoutPipe()
	if err != nil {
		return false, err
	}
	i, err := cmd.StdinPipe()
	if err != nil {
		return false, err
	}

	if _, err = i.Write([]byte(payload)); err != nil {
		return false, err
	}
	i.Close()

	if err = cmd.Start(); err != nil {
		return false, err
	}

	b, err := ioutil.ReadAll(o)
	if err != nil {
		return false, err
	}

	m, err := regexp.Match(`(?m)^\[GNUPG:\] GOODSIG`, b)
	if err != nil {
		return false, err
	}
	if !m {
		return false, nil
	}

	return true, nil
}
