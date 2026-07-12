package cli

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/DishanRajapaksha/industrial-cli-kit/command"
	"github.com/DishanRajapaksha/industrial-cli-kit/completion"
)

func TestRegistryMatchesDispatcher(t *testing.T) {
	dispatched := []string{
		"endpoints", "status", "namespaces", "browse", "tui", "attributes", "read", "write",
		"monitor", "watch", "alarms", "test-connection", "init-config", "validate-config",
		"completions", "help", "version",
	}
	registered := map[string]bool{}
	for _, registeredCommand := range cliRegistry.Commands {
		if registered[registeredCommand.Name] {
			t.Fatalf("duplicate registry command %q", registeredCommand.Name)
		}
		registered[registeredCommand.Name] = true
	}
	for _, name := range dispatched {
		if !registered[name] {
			t.Errorf("dispatcher command %q is not registered", name)
		}
	}
	for name := range registered {
		found := false
		for _, candidate := range dispatched {
			if candidate == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("registered command %q is not dispatched", name)
		}
	}
}

func TestRegistryGlobalFlagsMatchNormalizer(t *testing.T) {
	for _, global := range cliRegistry.GlobalFlags {
		args := []string{"--" + global.Name}
		if global.TakesValue {
			args = append(args, "value")
		}
		args = append(args, "status")
		normalised, err := normaliseGlobalFlags(args)
		if err != nil {
			t.Errorf("registered global flag --%s is rejected: %v", global.Name, err)
			continue
		}
		if len(normalised) == 0 || normalised[0] != "status" {
			t.Errorf("normalising --%s produced %v", global.Name, normalised)
		}
	}
}

func TestRegistryAppliesCommandGlobalPolicies(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []string
	}{
		{
			name: "endpoint discovery drops session security globals",
			args: []string{
				"--config", "site.yaml", "--profile", "local", "--endpoint", "opc.tcp://host:4840",
				"--policy", "Basic256Sha256", "--format", "json", "--verbose", "endpoints",
			},
			want: []string{
				"endpoints", "--config", "site.yaml", "--profile", "local", "--endpoint", "opc.tcp://host:4840",
				"--format", "json", "--verbose",
			},
		},
		{
			name: "validation keeps only local configuration globals",
			args: []string{
				"--config", "site.yaml", "--profile", "local", "--endpoint", "opc.tcp://host:4840",
				"--policy", "Basic256Sha256", "--format", "json", "--debug", "validate-config",
			},
			want: []string{"validate-config", "--config", "site.yaml", "--profile", "local", "--debug"},
		},
		{
			name: "diagnostics drops output format",
			args: []string{
				"--endpoint", "opc.tcp://host:4840", "--policy", "Basic256Sha256",
				"--format", "json", "--verbose", "test-connection",
			},
			want: []string{
				"test-connection", "--endpoint", "opc.tcp://host:4840", "--policy", "Basic256Sha256", "--verbose",
			},
		},
		{
			name: "init config rejects inherited globals by omission",
			args: []string{"--config", "site.yaml", "--verbose", "init-config", "--output", "new.yaml"},
			want: []string{"init-config", "--output", "new.yaml"},
		},
		{
			name: "completion shell stays before flags",
			args: []string{"--config", "site.yaml", "completions", "bash"},
			want: []string{"completions", "bash"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := normaliseGlobalFlags(test.args)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("normaliseGlobalFlags() = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestRegistryCapturesWriteSafetyAndLocalCommands(t *testing.T) {
	write := registryCommand(t, "write")
	assertFlag(t, write.Flags, "yes", false)
	assertFlag(t, write.Flags, "dry-run", false)
	assertFlag(t, write.Flags, "item", true)

	completions := registryCommand(t, "completions")
	if completions.LeadingArgs != 1 {
		t.Fatalf("completions LeadingArgs=%d, want 1", completions.LeadingArgs)
	}
	if completions.GlobalFlags == nil || len(completions.GlobalFlags) != 0 {
		t.Fatalf("completions must explicitly reject global flags: %#v", completions.GlobalFlags)
	}
	for _, name := range []string{"init-config", "help", "version"} {
		registered := registryCommand(t, name)
		if registered.GlobalFlags == nil || len(registered.GlobalFlags) != 0 {
			t.Fatalf("%s must explicitly reject global flags: %#v", name, registered.GlobalFlags)
		}
	}
}

func TestGeneratedCompletionsApplyPoliciesAndSafetyFlags(t *testing.T) {
	var out bytes.Buffer
	if err := completion.Write(&out, "bash", cliRegistry); err != nil {
		t.Fatal(err)
	}
	script := out.String()

	assertCaseContains(t, script, "endpoints", "--config", "--debug", "--endpoint", "--format", "--profile", "--timeout", "--verbose")
	assertCaseOmits(t, script, "endpoints", "--policy", "--mode", "--username", "--password", "--cert", "--key")

	assertCaseContains(t, script, "validate-config", "--config", "--debug", "--profile", "--verbose")
	assertCaseOmits(t, script, "validate-config", "--endpoint", "--policy", "--format")

	assertCaseContains(t, script, "test-connection", "--endpoint", "--policy", "--mode", "--timeout", "--verbose")
	assertCaseOmits(t, script, "test-connection", "--format")

	assertCaseContains(t, script, "init-config", "--force", "--output")
	assertCaseOmits(t, script, "init-config", "--config", "--profile", "--verbose")
	assertCaseContains(t, script, "write", "--yes", "--dry-run", "--item", "--value")

	for _, name := range []string{"completions", "help", "version"} {
		assertCaseOmits(t, script, name, "--config", "--endpoint", "--verbose")
	}
	if !strings.Contains(script, "complete -F _opc_ua_cli_completion opc-ua-cli") {
		t.Fatalf("completion is not registered for opc-ua-cli: %s", script)
	}
}

func registryCommand(t *testing.T, name string) command.Command {
	t.Helper()
	for _, registered := range cliRegistry.Commands {
		if registered.Name == name {
			return registered
		}
	}
	t.Fatalf("registry command %q not found", name)
	return command.Command{}
}

func assertFlag(t *testing.T, flags []command.Flag, name string, takesValue bool) {
	t.Helper()
	for _, flag := range flags {
		if flag.Name == name {
			if flag.TakesValue != takesValue {
				t.Fatalf("flag --%s TakesValue=%v, want %v", name, flag.TakesValue, takesValue)
			}
			return
		}
	}
	t.Fatalf("flag --%s not found", name)
}

func assertCaseContains(t *testing.T, script, name string, values ...string) {
	t.Helper()
	line := bashCaseLine(t, script, name)
	for _, value := range values {
		if !strings.Contains(line, value) {
			t.Errorf("%s completion is missing %q: %s", name, value, line)
		}
	}
}

func assertCaseOmits(t *testing.T, script, name string, values ...string) {
	t.Helper()
	line := bashCaseLine(t, script, name)
	for _, value := range values {
		if strings.Contains(line, value) {
			t.Errorf("%s completion unexpectedly includes %q: %s", name, value, line)
		}
	}
}

func bashCaseLine(t *testing.T, script, name string) string {
	t.Helper()
	prefix := "    " + name + ") words="
	for _, line := range strings.Split(script, "\n") {
		if strings.HasPrefix(line, prefix) {
			return line
		}
	}
	t.Fatalf("completion case for %q not found", name)
	return ""
}
