# Ambiente de desenvolvimento

O setup é o mesmo em macOS, Linux e Windows: o [mise](https://mise.jdx.dev) instala as
versões exatas das ferramentas (definidas em [`mise.toml`](mise.toml)) e o `make` orquestra o resto.

| Ferramenta     | Como é instalada                                   |
|----------------|----------------------------------------------------|
| Go             | mise (`mise.toml`)                                 |
| golang-migrate | mise (`mise.toml`)                                 |
| air            | mise (`mise.toml`)                                 |
| swag           | mise (`mise.toml`)                                 |
| mockgen        | dependência `tool` do `go.mod` (`go tool mockgen`) |
| Docker         | Docker Desktop / Docker Engine                     |

## 1. Pré-requisitos por sistema operacional

### macOS

```bash
brew install mise
echo 'eval "$(mise activate zsh)"' >> ~/.zshrc && exec zsh
```

O `make` e o `git` já vêm com as Command Line Tools do Xcode (`xcode-select --install`).

**Docker via Homebrew + Colima (sem Docker Desktop):** a fórmula `docker` instala só o CLI.
Instale os plugins e aponte o CLI para eles:

```bash
brew install docker docker-compose docker-buildx colima
```

Adicione ao `~/.docker/config.json`:

```json
"cliPluginsExtraDirs": ["/opt/homebrew/lib/docker/cli-plugins"]
```

Sem o `docker-buildx`, o build cai no builder legado e falha com `"" is an invalid OS component`.
O Elastic Stack precisa de memória: o padrão do Colima (2 GiB) faz o kernel matar o
Elasticsearch (exit 137). Use pelo menos 6 GiB:

```bash
colima start --memory 8 --cpus 4
```

### Linux

```bash
curl https://mise.run | sh
echo 'eval "$(~/.local/bin/mise activate bash)"' >> ~/.bashrc && exec bash
```

Instale `make` e `git` pelo gerenciador da distro (ex.: `sudo apt install make git`).

### Windows (nativo)

Execute no PowerShell:

```powershell
winget install --id Git.Git           # Git for Windows (inclui o Git Bash)
winget install --id jdx.mise          # gerenciador de versões
winget install --id ezwinports.make   # GNU make
winget install --id Docker.DockerDesktop
```

> Com Scoop: `scoop install git mise make`. Com Chocolatey: `choco install git mise make`.

Adicione os *shims* do mise ao `PATH` do usuário. Assim, Go, air e migrate funcionam em
qualquer terminal e também em IDEs (VS Code, GoLand):

```powershell
[Environment]::SetEnvironmentVariable("Path", "$env:LOCALAPPDATA\mise\shims;" + [Environment]::GetEnvironmentVariable("Path", "User"), "User")
```

Feche e reabra o terminal. **Use o Git Bash** para rodar os comandos `make`: o `Makefile` e
o git hook usam sh POSIX, que o Git for Windows já fornece.

## 2. Setup do projeto (todos os SOs)

```bash
git clone <url-do-repositorio> && cd go-payment-apm
make setup
```

O `make setup`:

1. cria o `.env` a partir do `.env.example`;
2. roda `mise install` (Go, migrate, air, swag) e `go mod download`;
3. ativa o pre-commit hook (`git config core.hooksPath .githooks`).

Depois disso: `make deps-up`, `make migrate-up` e `make run`. Rode `make` sem argumentos
para ver todos os comandos.

## 3. Formatação e pre-commit hook

- `make fmt` formata todo o código com `go fmt ./...` e falha se algum arquivo tiver erro de sintaxe.
- `make fmt-check` só verifica, sem alterar nada. É a versão para o CI.

O hook [`.githooks/pre-commit`](.githooks/pre-commit) é versionado no repositório e roda
`make fmt` a cada commit que inclua arquivos `.go`. O commit é **abortado** quando:

- o `make fmt` falha (por exemplo, por um erro de sintaxe); ou
- o `go fmt` reformatou algum arquivo. Nesse caso, revise com `git diff`, rode `git add` e
  faça o commit de novo.

Como o hook é ativado por `core.hooksPath`, qualquer atualização dele chega a todos pelo
próprio `git pull`, sem copiar nada para `.git/hooks`. Para pular o hook pontualmente, use
`git commit --no-verify`.

## Migrando do asdf

O `.tool-versions` foi substituído pelo `mise.toml`. Para migrar, instale o mise como
descrito acima, rode `mise install` e remova os plugins do asdf (`asdf plugin remove golang`
etc.) se não forem mais usados em outros projetos.

## Fins de linha

O [`.gitattributes`](.gitattributes) força LF em todos os arquivos de texto. Sem isso, o Git
no Windows faria checkout com CRLF, o que quebraria o `Makefile` e os scripts shell. Também
faria o `go fmt` reescrever todos os arquivos `.go`, e o pre-commit barraria todo commit.
