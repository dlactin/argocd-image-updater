package git

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/argoproj-labs/argocd-image-updater/pkg/log"
)

// CommitOptions holds options for a git commit operation
type CommitOptions struct {
	// CommitMessageText holds a short commit message (-m option)
	CommitMessageText string
	// CommitMessagePath holds the path to a file to be used for the commit message (-F option)
	CommitMessagePath string
	// SigningKey holds a GnuPG key ID or path to Public SSH Key used to sign the commit with (-S option)
	SigningKey string
	// SignOff specifies whether to sign-off a commit (-s option)
	SignOff bool
}

// Commit perfoms a git commit for the given pathSpec to the currently checked
// out branch. If pathSpec is empty, or the special value "*", all pending
// changes will be commited. If message is not the empty string, it will be
// used as the commit message, otherwise a default commit message will be used.
// If signingKey is not the empty string, commit will be signed with the given
// GPG key.
func (m *nativeGitClient) Commit(pathSpec string, opts *CommitOptions) error {
	defaultCommitMsg := "Update parameters"
	args := []string{"commit"}

  log.Warnf("Debug Build - Dlactin")
  log.Warnf("Options: %v", opts)
	if pathSpec == "" || pathSpec == "*" {
		args = append(args, "-a")
	}
	if opts.SigningKey != "" {
		// Check if SiginingKey is a GPG key or Public SSH Key
		keyCheck, err := regexp.MatchString(".*pub$", opts.SigningKey)
		if err != nil {
			return fmt.Errorf("could not validate Signing Key as GPG or Public SSH Key: %v", err)
		}
		if keyCheck {
			args = append(args, "-S")
		} else {
			args = append(args, "-S", opts.SigningKey)
		}
	}
	if opts.SignOff {
		args = append(args, "-s")
	}
	if opts.CommitMessageText != "" {
		args = append(args, "-m", opts.CommitMessageText)
	} else if opts.CommitMessagePath != "" {
		args = append(args, "-F", opts.CommitMessagePath)
	} else {
		args = append(args, "-m", defaultCommitMsg)
	}

	out, err := m.runCmd(args...)
	if err != nil {
		log.Errorf(out)
		return err
	}

	return nil
}

// Branch creates a new target branch from a given source branch
func (m *nativeGitClient) Branch(sourceBranch string, targetBranch string) error {
	if sourceBranch != "" {
		_, err := m.runCmd("checkout", sourceBranch)
		if err != nil {
			return fmt.Errorf("could not checkout source branch: %v", err)
		}
	}

	_, err := m.runCmd("branch", targetBranch)
	if err != nil {
		return fmt.Errorf("could not create new branch: %v", err)
	}

	return nil
}

// Push pushes local changes to the remote branch. If force is true, will force
// the remote to accept the push.
func (m *nativeGitClient) Push(remote string, branch string, force bool) error {
	args := []string{"push"}
	if force {
		args = append(args, "-f")
	}
	args = append(args, remote, branch)
	err := m.runCredentialedCmd("git", args...)
	if err != nil {
		return fmt.Errorf("could not push %s to %s: %v", branch, remote, err)
	}
	return nil
}

// Add adds a path spec to the repository
func (m *nativeGitClient) Add(path string) error {
	return m.runCredentialedCmd("git", "add", path)
}

// SymRefToBranch retrieves the branch name a symbolic ref points to
func (m *nativeGitClient) SymRefToBranch(symRef string) (string, error) {
	output, err := m.runCmd("symbolic-ref", symRef)
	if err != nil {
		return "", fmt.Errorf("could not resolve symbolic ref '%s': %v", symRef, err)
	}
	if a := strings.SplitN(output, "refs/heads/", 2); len(a) == 2 {
		return a[1], nil
	}
	return "", fmt.Errorf("no symbolic ref named '%s' could be found", symRef)
}

// Config configures username and email address for the repository
func (m *nativeGitClient) Config(username string, email string) error {
	_, err := m.runCmd("config", "user.name", username)
	if err != nil {
		return fmt.Errorf("could not set git username: %v", err)
	}
	_, err = m.runCmd("config", "user.email", email)
	if err != nil {
		return fmt.Errorf("could not set git email: %v", err)
	}

	return nil
}

// SigningConfig configures commit signing
func (m *nativeGitClient) SigningConfig(signingkey string) error {
  // Check if SiginingKey is a GPG key or Public SSH Key
  keyCheck, err := regexp.MatchString(".*pub$", signingkey)
  if err != nil {
    return fmt.Errorf("could not validate Signing Key as GPG or Public SSH Key: %v", err)
  }
  if keyCheck {
    // Setting the GPG format to ssh
    log.Warnf("Setting GPG Format to SSH")
    _, err = m.runCmd("config", "gpg.format", "ssh")
    if err != nil {
      return fmt.Errorf("could not set gpg format to ssh: %v", err)
    }
    // Setting Public SSH Key as our signing key
    // SSH Keys can not currently be set via cli flag
    _, err = m.runCmd("config", "user.signingkey", signingkey)
    if err != nil {
      return fmt.Errorf("could not set git signing key: %v", err)
    }
  }

	return nil
}
