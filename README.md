# SPOTCIN (Equipe 3) - Streaming de Áudio Distribuído

## Descrição do Projeto
SPOTCIN é um sistema distribuído de streaming de áudio MP3/WAV que utiliza HTTP Streaming com requisições Range. A aplicação conta com um catálogo de áudios, onde é possível escolher um áudio e iniciar, pausar ou avançar até determinado momento.

## Como Rodar

### Pré-requisitos
- [Go](https://go.dev/dl/) 1.27.4 ou superior
- [Docker](https://docs.docker.com/get-docker/) com Docker Compose
- [Node.js](https://nodejs.org/) 20 ou superior (apenas para o frontend)

### 1. Clonar o projeto
```bash
git clone https://github.com/Pedro-glitch-370/SPOTCIN.git
cd SPOTCIN
```

### 2. Configurar as variáveis de ambiente
O backend lê o arquivo `backend/.env` ao iniciar. Crie-o a partir do modelo em `deploy/`:

Windows (PowerShell):
```powershell
copy deploy\.env.example backend\.env
```
Linux/macOS:
```bash
cp deploy/.env.example backend/.env
```

| Variável | Padrão | Descrição |
|---|---|---|
| `DATABASE_URL` | — (obrigatória) | Conexão com o PostgreSQL |
| `REDIS_ADDR` | `localhost:6379` | Endereço do Redis |
| `GARAGE_ENDPOINT` | `localhost:3900` | Endpoint S3 do Garage |
| `GARAGE_ACCESS_KEY` | `minioadmin` | Chave de acesso do Garage |
| `GARAGE_SECRET_KEY` | `minioadminminioadmin` | Chave secreta do Garage |
| `GARAGE_BUCKET` | `spotcin-audio` | Bucket dos áudios |
| `AUDIO_DIR` | `devseed/audio` | Pasta local de onde os áudios são lidos |

### 3. Subir a infraestrutura (PostgreSQL, Redis e Garage)
```bash
cd deploy
docker compose up -d
cd ..
```

### 4. Iniciar o servidor (porta 8080)
```bash
cd backend
go run ./cmd/server
```
Ao iniciar, o servidor conecta ao PostgreSQL, ao Redis e ao Garage, e indexa os áudios de `devseed/audio` no banco. Se alguma das dependências estiver fora do ar, o servidor encerra com erro.

### 5. Iniciar o frontend (porta 5173)
Em outro terminal, a partir da raiz do projeto:
```bash
cd frontend
npm install
npm run dev
```
O Vite redireciona as chamadas `/api` para `http://localhost:8080`.

## Entrega 1
Nesta fase, os metadados do catálogo ficam no **PostgreSQL**, e o Redis e o Garage já estão conectados (verificados no health-check). Os arquivos de áudio ainda são lidos da pasta local `backend/devseed/audio`; a migração deles para o bucket do Garage está planejada.

Para conferir a realização das tarefas da Entrega 1, inicie o servidor e, em outro terminal, execute os comandos abaixo.

> **Windows:** no PowerShell, `curl` é um apelido para outro comando; use `curl.exe`, como nos exemplos. O Python é chamado de `python` no Windows e de `python3` no Linux.

### Health-check
Confirmar que o processo está vivo:

Windows (PowerShell):
```powershell
curl.exe -i http://localhost:8080/health
```
Linux/macOS:
```bash
curl -i http://localhost:8080/health
```
A saída esperada é `{"status":"ok"}`.

Confirmar as dependências:

Windows (PowerShell):
```powershell
curl.exe -s http://localhost:8080/health/ready | python -m json.tool
```
Linux/macOS:
```bash
curl -s http://localhost:8080/health/ready | python3 -m json.tool
```
Saída esperada:
```json
{
    "status": "ok",
    "checks": {
        "garage": "ok",
        "postgres": "ok",
        "redis": "ok"
    }
}
```
Se alguma dependência falhar, a resposta é `503 Service Unavailable` com `"status": "degraded"` e o erro no campo correspondente.

### Catálogo
Windows (PowerShell):
```powershell
curl.exe -s http://localhost:8080/api/songs | python -m json.tool
```
Linux/macOS:
```bash
curl -s http://localhost:8080/api/songs | python3 -m json.tool
```
Exemplo de saída esperada:
```json
[
    {
        "id": "708ef970-d1aa-4c48-bd31-efe81facb5a6",
        "path": "kokiri_forest.wav",
        "title": "Kokiri Forest",
        "artist": "Unknown",
        "size": 10769278,
        "mod_time": "2026-11-17T16:33:09.270541-03:00",
        "mime_type": "audio/wav",
        "created_at": "2026-11-17T19:45:15.9978222-03:00"
    }
]
```

### Streaming com Range (resposta 206 + Content-Range)
Substitua `<ID>` por um `id` retornado pelo catálogo:

Windows (PowerShell):
```powershell
curl.exe -i -H "Range: bytes=0-1023" http://localhost:8080/api/stream/<ID>
```
Linux/macOS:
```bash
curl -i -H "Range: bytes=0-1023" http://localhost:8080/api/stream/<ID>
```
Exemplo de saída esperada:
```text
HTTP/1.1 206 Partial Content
Accept-Ranges: bytes
Access-Control-Allow-Headers: Content-Type, Authorization, Range
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
Access-Control-Allow-Origin: *
Access-Control-Expose-Headers: Content-Range, Accept-Ranges, Content-Length
Content-Length: 1024
Content-Range: bytes 0-1023/10769278
Content-Type: audio/wav
Last-Modified: Tue, 17 Nov 2026 19:33:09 GMT
Date: Tue, 17 Nov 2026 22:49:48 GMT
```

### Range inválido (resposta 416)
Windows (PowerShell):
```powershell
curl.exe -i -H "Range: bytes=999999999-" http://localhost:8080/api/stream/<ID>
```
Linux/macOS:
```bash
curl -i -H "Range: bytes=999999999-" http://localhost:8080/api/stream/<ID>
```
Exemplo de saída esperada:
```text
HTTP/1.1 416 Requested Range Not Satisfiable
Access-Control-Allow-Headers: Content-Type, Authorization, Range
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
Access-Control-Allow-Origin: *
Access-Control-Expose-Headers: Content-Range, Accept-Ranges, Content-Length
Content-Range: bytes */10769278
Content-Type: text/plain; charset=utf-8
X-Content-Type-Options: nosniff
Date: Tue, 17 Nov 2026 22:50:48 GMT
Content-Length: 33

invalid range: failed to overlap
```

## Entrega 2
Quanto aos logs operacionais, todas as requisições HTTP viraram linhas JSON no stdout do servidor. Além disso, os áudios passaram a viver no Garage, o que tornou o disco local apenas uma fonte de seed. Com isso, o backend tornou-se stateless de verdade. Por fim, foram implementadas as rotas de playback com o Redis e o cliente de terminal.

### Salvar Posição de Reprodução
Substitua `<ID>` por um `id` retornado pelo catálogo:

Windows (PowerShell):
```powershell
curl.exe -sS -X PUT -H "X-User-ID: test-1" -H "Content-Type: application/json" `
  --data-raw "{\"posicao_seg\": 60}" http://localhost:8080/api/playback/<ID>
```
Linux/macOS:
```bash
curl -sS -X PUT -H "X-User-ID: test-1" -H "Content-Type: application/json" -d '{"posicao_seg": 60}' http://localhost:8080/api/playback/<ID>
```
Exemplo de saída esperada no terminal do servidor:
```json
{"time":"2026-11-21T11:25:03.3969538-03:00","level":"INFO","msg":"http_request","method":"PUT","path":"/api/playback/feadc394-bc6f-40e1-835c-dd9fbdda4f15","status":204,"duration_ms":7,"bytes":0,"request_id":"ec5db4fd-4103-4e84-ac93-ebd6328d2e21"}
```

### Resgatar Última Posição de Reprodução
Substitua `<ID>` por um `id` retornado pelo catálogo:
Windows (PowerShell):
```powershell
curl.exe -sS -H "X-User-ID: test-1" http://localhost:8080/api/playback/<ID>
```
Linux/macOS:
```bash
curl -sS -H "X-User-ID: test-1" http://localhost:8080/api/playback/<ID>
```
O comando irá retornar imediatamente:
```json
{"track_id":"feadc394-bc6f-40e1-835c-dd9fbdda4f15","posicao_seg":60,"atualizado_em":"2026-11-21T14:25:03.3941125Z"}
```
Exemplo de saída esperada no terminal do servidor:
```json
{"time":"2026-11-21T11:26:19.3842379-03:00","level":"INFO","msg":"http_request","method":"GET","path":"/api/playback/feadc394-bc6f-40e1-835c-dd9fbdda4f15","status":200,"duration_ms":1,"bytes":116,"request_id":"4f89f13b-216f-4479-89e2-f769396ca759"}
```

### Cliente de Terminal
O cliente consome o stream com requisições Range e um buffer de reprodução, sem baixar o arquivo completo antes de começar. Com o servidor rodando, em outro terminal:
```bash
cd backend
go run ./cmd/client
```
Sem a flag `-song`, o cliente lista o catálogo e pede para escolher uma música pelo número ou pelo `id`. Se houver posição salva no Redis para o usuário, a reprodução retoma desse ponto.

Exemplo com flags:
```bash
go run ./cmd/client -song <ID> -user test-1 -start 30 -min 65536 -max 524288 -jitter 100ms -drop-after 1048576
```

| Flag | Padrão | Descrição |
|---|---|---|
| `-base` | `http://localhost:8080` | URL do backend |
| `-song` | — | `id` da música (se vazio, lista o catálogo) |
| `-user` | UUID v4 gerado | Valor do cabeçalho `X-User-ID` |
| `-start` | posição salva no Redis | Posição inicial em segundos |
| `-min` | `65536` | Nível mínimo do buffer (bytes) para iniciar a reprodução |
| `-max` | `524288` | Nível máximo do buffer (bytes); acima disso o download pausa |
| `-jitter` | `50ms` | Oscilação aleatória simulada no consumo |
| `-drop-after` | `0` (desligado) | Fecha a conexão após N bytes, para simular quedas |

#### Comandos durante a reprodução
Com o cliente tocando, digite um comando e pressione Enter:

| Comando | Ação |
|---|---|
| `p` | Pausa ou retoma a reprodução |
| `a <seg>` | Avança `<seg>` segundos |
| `v <seg>` | Volta `<seg>` segundos |
| `i <seg>` | Vai para a posição `<seg>` da faixa |
| `q` | Salva a posição e encerra |

Ao avançar, voltar ou ir para uma posição, o cliente descarta o buffer, converte a posição em offset de bytes e abre uma nova requisição `Range: bytes=<offset>-`. A nova posição também é salva no Redis. Exemplo de saída:
```text
posição: 45.00s (Range: bytes=7938000-)
2026/11/22 17:21:24 conexão #1: resume em 7938000 bytes (45.00s)
2026/11/22 17:21:24 stream aberto: status=206 range=bytes=7938000-
```

#### Reconexão
Se a conexão cair (ou for fechada com `-drop-after`), o cliente reconecta a partir do último byte recebido, e não da posição do player, de modo que os bytes que já estão no buffer não são baixados de novo:
```text
downloader: drop-after atingido (1024 KB) — fechando socket
downloader: reconectando após drop (próximo offset=1048576)
conexão #2: resume em 1048576 bytes (5.94s)
```

### Validação com N clientes
Para verificar a concorrência, faça:
```bash
cd SPOTCIN
./scripts/load.sh
```
O script dispara 4 clientes em paralelo (2 por música, com -user e drops distintos). Os logs do servidor mostram que eles possuem request_id distintos, status:206, range diferentes e durações sobrepostas.

**OBS**: O cliente ainda não toca áudio de verdade. Ele apenas simula a reprodução consumindo do buffer conforme a taxa do bit rate real junto com um jitter. O objetivo foi tornar o buffer e os requisitos da entrega observáveis no terminal, pois o produto final usará <audio> no frontend.

## Framework de Streaming
Foram consideradas duas opções sugeridas: HTTP Streaming e WebSockets.

A preferência foi por HTTP Streaming com Range, pois é entendido nativamente pela tag `<audio>`, utilizada pelo React no frontend, e facilmente atendido pelo pacote `net/http` do Go no backend (`http.ServeContent` já trata os cabeçalhos `Range`, `Accept-Ranges` e `Content-Range` e as respostas 206 e 416).

Além disso, o uso de HTTP Streaming deixa os servidores stateless, ou seja, torna cada request de Range independente. Isso permite escalar melhor a aplicação em várias máquinas atrás de um load balancer.

Por fim, a aplicação usa áudios já prontos ao invés de realizar broadcast em tempo real, o que elimina a necessidade de WebSockets.

## Tecnologias Escolhidas
- **Frontend:** Vite + React em TypeScript
- **Backend:** Go, usando `net/http` da biblioteca padrão e GORM para o banco
- **Metadados:** PostgreSQL 16
- **Posições de reprodução:** Redis 7
- **Binários de áudio:** Garage (object storage compatível com S3, acessado pelo cliente `minio-go`)
- **Load Balancer:** Nginx (planejado)
- **Infraestrutura local:** Docker Compose

## Arquitetura
```text
                    ┌───────────────┐
  Browser (React) ─>│ Load Balancer │  (Nginx, planejado)
                    └───────┬───────┘
              ┌─────────────┼─────────────┐
              v             v             v
         ┌────────┐    ┌────────┐    ┌────────┐
         │ Go API │    │ Go API │    │ Go API │  <- stateless, N réplicas
         └───┬────┘    └───┬────┘    └───┬────┘
             └─────────────┼─────────────┘
          ┌────────────────┼────────────────┐
          v                v                v
        Redis          PostgreSQL         Garage (S3)
      (posição)       (metadados)     (binários de áudio)
```
Como nenhuma instância da API guarda estado em memória, qualquer réplica pode atender qualquer requisição, inclusive requisições Range diferentes de uma mesma faixa.

## Estrutura de Pastas
```text
SPOTCIN/
├── backend/                      # Servidor em Go (stateless)
│   ├── cmd/
│   │   └── server/
│   │       └── main.go           # Conecta às dependências, indexa os áudios e sobe o HTTP
│   ├── internal/
│   │   ├── database/             # Conexões: PostgreSQL (db.go), Redis (redis.go), Garage (garage.go)
│   │   ├── handlers/             # Camada HTTP (health, catálogo, stream)
│   │   ├── middleware/           # CORS e logging
│   │   ├── models/               # Structs de domínio (Song)
│   │   ├── routes/               # Registro de rotas
│   │   └── services/             # Regras de negócio (scanner de áudios, consultas)
│   ├── devseed/
│   │   └── audio/                # Áudios de exemplo
│   ├── go.mod
│   └── go.sum
├── frontend/                     # Vite + React (TypeScript)
│   ├── public/
│   ├── src/
│   ├── index.html
│   └── vite.config.ts
├── deploy/                       # Infraestrutura
│   ├── docker-compose.yml        # PostgreSQL, Redis e Garage
│   ├── garage.toml               # Configuração do Garage
│   ├── .env.example              # Modelo de variáveis de ambiente
│   └── lb/
│       └── nginx.conf            # Load Balancer (planejado)
└── README.md
```

## Estrutura de Dados e Catálogo
O sistema lida com três tipos de dados, cada um com um tipo de armazenamento distinto.

### 1. Metadados do catálogo (PostgreSQL)
Os metadados têm estrutura fixa e são consultados por atributos, por isso ficam em um banco relacional. A tabela `songs` é criada automaticamente pelo GORM a partir do modelo `Song` (`backend/internal/models/song.go`):

| Campo | Tipo | Descrição |
|---|---|---|
| `id` | texto (UUID v4) | Identificador da faixa |
| `path` | texto, único | Nome do arquivo dentro da pasta de áudios |
| `title` | texto | Título (gerado a partir do nome do arquivo) |
| `artist` | texto | Artista |
| `album` | texto | Álbum (opcional) |
| `duration` | inteiro | Duração em segundos (campo previsto; o cálculo ainda não foi implementado) |
| `size` | inteiro | Tamanho em bytes |
| `mod_time` | data/hora | Última modificação do arquivo |
| `mime_type` | texto | `audio/mpeg` (MP3) ou `audio/wav` (WAV) |
| `created_at` | data/hora | Data de indexação |

O catálogo é preenchido pelo scanner (`backend/internal/services/scanner.go`) ao iniciar o servidor. Ele percorre a pasta de áudios, insere arquivos novos, atualiza os que mudaram de tamanho ou data e remove do banco os que sumiram do disco.

### 2. Binários de áudio (Garage)
Os arquivos de áudio são grandes, imutáveis e lidos por intervalo (Range). Por isso, o armazenamento escolhido é um object storage compatível com S3, que oferece leitura por intervalo nativamente. O **Garage** sobe pelo Docker Compose com o bucket `spotcin-audio`, e o servidor já conecta a ele na inicialização.

Na Entrega 1 os áudios ainda são servidos da pasta local `backend/devseed/audio`. A organização planejada no bucket é:
```text
spotcin-audio/
  mp3/<uuid>.mp3
  wav/<uuid>.wav
```
Onde `<uuid>` é o mesmo `id` do registro correspondente no catálogo.

### 3. Posições de reprodução (Redis)
As posições são escritas com frequência (a cada avanço de reprodução), exigem baixa latência e têm vida curta, por isso ficam no **Redis**. A leitura e a escrita são feitas pelas rotas `GET` e `PUT /api/playback/{id}`, identificando o usuário pelo cabeçalho `X-User-ID`.

A chave tem o formato `playback:<user_id>:<track_id>` e o valor tem o formato `{"posicao_seg": 123.4, "atualizado_em": "..."}`, com TTL configurável.

## Validações
- **Formato do arquivo:** o scanner só indexa arquivos com extensão de áudio e confirma o tipo lendo os primeiros bytes do arquivo (assinatura do formato), não apenas a extensão. Arquivos que não são áudio são ignorados e registrados no log.
- **Intervalo solicitado:** requisições `Range` fora do tamanho do arquivo recebem `416 Range Not Satisfiable` com `Content-Range: bytes */<tamanho>`.
- **Faixa inexistente:** um `id` que não está no catálogo recebe `404 Not Found`.
- **Path traversal:** antes de abrir o arquivo, o servidor confirma que o caminho resolvido está dentro da pasta de áudios; caso contrário, responde `403 Forbidden`.

## Estratégia do Buffer
A reprodução de áudio no cliente é feita pelo elemento `<audio>` nativo do navegador, apoiado pelo protocolo HTTP definido no backend. Como o navegador já implementa, de forma nativa e otimizada, o controle de buffer, cabe ao frontend apenas observar e reagir aos eventos do `<audio>`.

### Tamanho dos Chunks
O navegador decide, por conta própria, quantos bytes solicitar a cada requisição. Para isso, ele faz requisições Range ao backend. O backend responde 206 Partial Content com o trecho solicitado.

### Nível Mínimo para Reprodução
O início da reprodução é definido pelos eventos do `<audio>` ao invés de um número fixo de segundos. O navegador começa a reproduzir assim que recebe dados suficientes para decodificar (evento `canplay`).

### Nível Máximo do Buffer
O navegador, de forma nativa, para de baixar quando o buffer do `<audio>` fica cheio. Logo, a determinação de um teto é opcional. Pode ser definido um teto de 30 a 40 s para calibrar o desempenho da reprodução.

### Estratégia contra Jitter
O `<audio>` absorve o jitter por meio do buffer à frente do ponto de reprodução. Enquanto houver dados no buffer, a reprodução continua mesmo que a rede oscile. Se o buffer esvaziar, o `<audio>` dispara o evento `waiting` e pausa a reprodução; o cliente exibe um indicador de carregamento até o evento `playing`.

### Pausar e Avançar na Faixa
Para pausar e retomar, são usados os métodos nativos `.pause()` e `.play()` do elemento `<audio>`.
Para avançar, o frontend altera o valor de `.currentTime` e o navegador dispara uma requisição Range correspondente ao backend. O backend responde 206 com o trecho e a reprodução retoma do novo ponto. Com isso, o frontend não precisa construir a requisição Range manualmente.

## Protocolo de Transmissão
O protocolo de transmissão escolhido é HTTP/1.1 com suporte a requisições Range.

### Rotas
| Método | Rota | Descrição | Status |
|---|---|---|---|
| GET | `/health` | Processo vivo | Implementada |
| GET | `/health/ready` | Estado do PostgreSQL, Redis e Garage | Implementada |
| GET | `/api/songs` | Catálogo | Implementada |
| GET | `/api/stream/{id}` | Áudio, com suporte a Range | Implementada |
| GET | `/api/songs/{id}` | Metadados de uma faixa | Implementada |
| GET | `/api/playback/{id}` | Última posição de reprodução | Implementada |
| PUT | `/api/playback/{id}` | Salva a posição de reprodução | Implementada |

**OBS:** `{id}` é o UUID v4 da faixa.

### GET /health
```text
HTTP/1.1 200 OK
Content-Type: application/json
Cache-Control: no-store

{"status":"ok"}
```

### GET /health/ready
Response 200 (todas as dependências ok):
```json
{"status":"ok","checks":{"garage":"ok","postgres":"ok","redis":"ok"}}
```

Response 503 (alguma dependência indisponível):
```json
{"status":"degraded","checks":{"garage":"ok","postgres":"ok","redis":"fail: dial tcp [::1]:6379: connect: connection refused"}}
```

### GET /api/songs
```text
HTTP/1.1 200 OK
Content-Type: application/json
```
O corpo é uma lista JSON de faixas, com os campos da tabela `songs` (ver exemplo em [Catálogo](#catálogo)). Em caso de erro no banco, a resposta é `500 Internal Server Error`.

### GET /api/stream/{id}
Sem Range: `200 OK` com o arquivo completo.
```text
HTTP/1.1 200 OK
Accept-Ranges: bytes
Content-Type: audio/wav
Content-Length: 10769278
```

Com Range satisfazível:
```text
GET /api/stream/{id} HTTP/1.1
Range: bytes=1048576-2097151
```
```text
HTTP/1.1 206 Partial Content
Accept-Ranges: bytes
Content-Type: audio/wav
Content-Range: bytes 1048576-2097151/10769278
Content-Length: 1048576
```

Com Range não satisfazível:
```text
HTTP/1.1 416 Requested Range Not Satisfiable
Content-Range: bytes */10769278
```

Faixa inexistente:
```text
HTTP/1.1 404 Not Found
```

### GET /api/songs/{id}
Response 200:
```json
{
  "id": "0d36235f-93d1-47c1-81eb-e244f502bfc2",
  "title": "Kokiri Forest",
  "artist": "Unknown",
  "duration": 120,
  "size": 10769278,
  "mime_type": "audio/wav"
}
```
Response 404: `{"erro":"track_nao_encontrada"}`

### GET /api/playback/{id}
Response 200:
```json
{
  "track_id": "0d36235f-93d1-47c1-81eb-e244f502bfc2",
  "posicao_seg": 123.4,
  "atualizado_em": "2026-11-15T14:32:00Z"
}
```
Response 404: `{"erro":"posicao_nao_encontrada"}`

### PUT /api/playback/{id}
Request:
```json
{"posicao_seg": 123.4}
```
- `204 No Content`: posição salva.
- `422 Unprocessable Entity`: `{"erro":"posicao_invalida"}` (posição negativa ou ausente).
- `404 Not Found`: `{"erro":"track_nao_encontrada"}`.

## Autores
- [@adnalisia](https://github.com/adnalisia)
- [@brunoslm](https://github.com/brunoslm)
- [@Pedro-glitch-370](https://github.com/Pedro-glitch-370)
- [@Valci2](https://github.com/Valci2)
- [@Vinicius2703](https://github.com/Vinicius2703)
