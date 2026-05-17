package cli

import (
	"errors"
	"fmt"
	"strings"
)

func (a *App) completions(args []string) error {
	if len(args) != 1 {
		return errors.New("usage: opc-ua-cli completions <bash|zsh>")
	}
	switch strings.ToLower(strings.TrimSpace(args[0])) {
	case "bash":
		_, err := fmt.Fprint(a.out, bashCompletionScript)
		return err
	case "zsh":
		_, err := fmt.Fprint(a.out, zshCompletionScript)
		return err
	default:
		return errors.New("unsupported shell, expected bash or zsh")
	}
}

const completionSubcommands = "endpoints status namespaces browse read write monitor watch alarms test-connection init-config completions help"
const completionCommonFlags = "--config --profile --endpoint --policy --mode --username --password --cert --key --timeout --format"

const bashCompletionScript = `#!/usr/bin/env bash
_opc_ua_cli_completions() {
  local cur prev words cword
  _init_completion || return

  local subcommands="` + completionSubcommands + `"
  local common_flags="` + completionCommonFlags + `"

  if [[ ${cword} -eq 1 ]]; then
    COMPREPLY=( $(compgen -W "${subcommands}" -- "${cur}") )
    return
  fi

  case "${words[1]}" in
    read)
      COMPREPLY=( $(compgen -W "${common_flags} --node --nodes" -- "${cur}") )
      ;;
    write)
      COMPREPLY=( $(compgen -W "${common_flags} --node --type --value --item --dry-run --yes" -- "${cur}") )
      ;;
    browse)
      COMPREPLY=( $(compgen -W "${common_flags} --node --depth" -- "${cur}") )
      ;;
    monitor|watch)
      COMPREPLY=( $(compgen -W "${common_flags} --node --interval --duration" -- "${cur}") )
      ;;
    alarms)
      COMPREPLY=( $(compgen -W "${common_flags} --node --interval --duration --min-severity" -- "${cur}") )
      ;;
    endpoints|status)
      COMPREPLY=( $(compgen -W "--config --profile --endpoint --timeout --format" -- "${cur}") )
      ;;
    namespaces|test-connection)
      COMPREPLY=( $(compgen -W "${common_flags}" -- "${cur}") )
      ;;
    init-config)
      COMPREPLY=( $(compgen -W "--output --force" -- "${cur}") )
      ;;
    completions)
      COMPREPLY=( $(compgen -W "bash zsh" -- "${cur}") )
      ;;
    *)
      COMPREPLY=( $(compgen -W "${common_flags}" -- "${cur}") )
      ;;
  esac
}
complete -F _opc_ua_cli_completions opc-ua-cli
`

const zshCompletionScript = `#compdef opc-ua-cli
_opc_ua_cli_completions() {
  local -a subcommands
  subcommands=(
    'endpoints:List server endpoints'
    'status:List server endpoints'
    'namespaces:List namespace indexes and URIs'
    'browse:Browse child nodes'
    'read:Read node values'
    'write:Write node values'
    'monitor:Subscribe to data changes'
    'watch:Poll node values'
    'alarms:Subscribe to alarms/events'
    'test-connection:Run connection diagnostics'
    'init-config:Write starter YAML config'
    'completions:Generate shell completion scripts'
  )

  local -a common_flags
  common_flags=(
    '--config[YAML config file]:config file:_files'
    '--profile[Config profile name]:profile:'
    '--endpoint[OPC UA endpoint URL]:endpoint:'
    '--policy[Security policy]:policy:'
    '--mode[Security mode]:mode:'
    '--username[Username]:username:'
    '--password[Password]:password:'
    '--cert[Client certificate file]:cert:_files'
    '--key[Client private key file]:key:_files'
    '--timeout[Request timeout]:duration:'
    '--format[Output format]:format:(table text json jsonl)'
  )

  if (( CURRENT == 2 )); then
    _describe 'subcommand' subcommands
    return
  fi

  case $words[2] in
    read)
      _arguments $common_flags '--node[node id]:node:' '--nodes[file with node ids]:file:_files'
      ;;
    write)
      _arguments $common_flags '--node[node id]:node:' '--type[value type]:type:' '--value[value]:value:' '--item[node:type:value]:item:' '--dry-run[dry run]' '--yes[skip confirmation]'
      ;;
    browse)
      _arguments $common_flags '--node[root node id]:node:' '--depth[browse depth]:depth:'
      ;;
    monitor|watch)
      _arguments $common_flags '--node[node id]:node:' '--interval[poll/subscription interval]:interval:' '--duration[stop after duration]:duration:'
      ;;
    alarms)
      _arguments $common_flags '--node[event source node]:node:' '--interval[subscription interval]:interval:' '--duration[stop after duration]:duration:' '--min-severity[min severity]:severity:'
      ;;
    endpoints|status)
      _arguments '--config[YAML config file]:config file:_files' '--profile[Config profile name]:profile:' '--endpoint[OPC UA endpoint URL]:endpoint:' '--timeout[Request timeout]:duration:' '--format[Output format]:format:(table json)'
      ;;
    namespaces|test-connection)
      _arguments $common_flags
      ;;
    init-config)
      _arguments '--output[output YAML config file]:file:_files' '--force[overwrite output file]'
      ;;
    completions)
      _arguments '1:shell:(bash zsh)'
      ;;
  esac
}
_opc_ua_cli_completions "$@"
`
