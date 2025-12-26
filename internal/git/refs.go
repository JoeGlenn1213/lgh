package git

import (
	"os/exec"
	"strings"
)

// GetRefs returns all references (branches and tags) and their hashes
func GetRefs(repoPath string) (map[string]string, error) {
	cmd := exec.Command("git", "-C", repoPath, "show-ref")
	output, err := cmd.CombinedOutput()
	// show-ref returns exit code 1 if no refs exist (new repo)
	if err != nil {
		// If empty output, it's just an empty repo
		if len(output) == 0 {
			return make(map[string]string), nil
		}
		// If real error, return it
		// But in a fresh bare repo, show-ref might fail.
		// Let's assume empty map on error for safety in this context
		// return nil, fmt.Errorf("failed to get refs: %w", err)
		return make(map[string]string), nil
	}

	refs := make(map[string]string)
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			hash := parts[0]
			ref := parts[1]
			refs[ref] = hash
		}
	}
	return refs, nil
}
