# SPOTCIN (Equipe 3)

## Descrição do Projeto
SPOTCIN é um sistema distribuído de streaming de áudio MP3/WAV que utiliza HTTP Streaming.

## Como Rodar
Clone o projeto:
```bash
  git clone https://github.com/Pedro-glitch-370/SPOTCIN.git
```

Acesse o diretório do projeto:
```bash
  cd SPOTCIN
```

Para iniciar o servidor:
- Entre na pasta /backend:
```
cd backend
```
- Inicialize com:
```
go run .
```

Para iniciar a aplicação na visão do cliente:
- Entre na pasta /frontend:
```
cd frontend
```
- Instale as dependências:
```
npm install
```
- Inicialize com:
```
npm run dev
```

## Framework de Streaming
Foram consideradas duas opções sugeridas: HTTP Streaming e WebSockets.

A preferência foi por HTTP Streaming com Range, pois é entendido nativamente pela tag <audio>, utilizada pelo React no frontend, e facilmente entendiido pelo pacote net/http do Go no backend.

Além disso, o uso de HTTP Streaming deixa os servidores stateless, ou seja, torna cada request de Range independente. Isso permite escalar melhor a aplicação em várias máquinas.

Por fim, a aplicação usa áudios já prontos ao invés de realizar broadcast em tempo real, o que elimina a necessidade de WebSockets.

## Tecnologias Escolhidas
A aplicação conta com:
- Frontend em Vite-React em Typescript
- Backend em Golang com foco em net/http
- Nginx como Load Balancer HTTP 
- Armazenamento de dados em MinIO, PostgreSQL e Redis

## Estrutura Principal de Pastas
```text
SPOTCIN/
├── backend/                     # Servidor em Golang (stateless)
│   ├── cmd/
│   │   └── server/              # Binário compilável e replicável
│   │       └── main.go          # Lê config, monta dependências, chama internal/, etc.
│   ├── internal/                # Lógica privado do Go (não exportável)
│   │   ├── config/              # Variáveis de ambiente
│   │   ├── database/
│   │   │   ├── postgres/        # Metadados
│   │   │   ├── redis/           # Posições de Reprodução
│   │   │   ├── minio/           # Binários de áudio
│   │   │   └── migrations/      # SQL versionado
│   │   ├── handlers/            # Camada HTTP
│   │   ├── middleware/          # CORS, request-id, etc.
│   │   ├── models/              # Structs de domínio
│   │   ├── routes/              # Registro de rotas
│   │   ├── services/            # Regras de negócio
│   │   └── storage/             # Interfaces Go
│   ├── pkg/                     # Código reutilizável (exportável)
│   ├── devseed/                 # Dados carregados no MinIO/Postgres via script
│   │   └── audio/
│   │       ├── mp3/
│   │       └── wav/
│   ├── go.mod
│   └── go.sum
├── frontend/                    # Vite + React (TypeScript)
│   ├── public/
│   ├── src/
│   │   ├── assets/              # Arquivos estáticos
│   │   ├── components/          # Componentes TSX
│   │   ├── pages/               # Páginas da aplicação
│   │   ├── hooks/               # useAudio, usePlayback, etc.
│   │   ├── services/            # Chamadas HTTP
│   │   ├── types/               # Tipos TypeScript espelhando os DTOs
│   │   ├── utils/               # Funções suporte
│   │   ├── App.tsx
│   │   └── main.tsx
│   ├── index.html
│   └── vite.config.ts
├── tests/                       # Testes unitários e de integração
├── deploy/                      # Como o sistema roda
│   ├── docker-compose.yml
│   ├── .env.example             # Template de variáveis de ambiente
│   └── lb/                      # Load Balancer (Nginx)
│       └── nginx.conf           # Configuração do Nginx
├── .gitignore
└── README.md
```

## Estrutura de Dados e Catálogo
O sistema lida com três tipos de dados, cada um com um tipo de armazenamento distinto.

### 1. Binários de Áudio (MinIO)
Esses arquivos são grandes, imutáveis e lidos por intervalo (Range). Por isso, foi escolhido como forma de armazenamento o **MinIO**, que oferece nativamente a leitura por intervalo.

Estrutura do diretório /audio:
```text
/audio/
  mp3/
    <uuid>.mp3
  wav/
    <uuid>.wav
```
Onde <uuid> é o mesmo id do registro correspondente na tabela tracks.

O caminho relativo (ex.: mp3/<uuid>.mp3) é gravado na coluna object_key do banco. Já o caminho absoluto é montado em Go como /audio/ + object_key.

### 2. Metadados (PostgreSQL)
Os metadados podem ser representados por um modelo relacional. Portanto, foi escolhido o **PostgreSQL** para permitir consultas por atributos e consistência.

Modelo do catálogo:
```sql
CREATE TABLE tracks (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title       TEXT NOT NULL,
    artist      TEXT NOT NULL,
    album       TEXT,
    duration_ms INTEGER NOT NULL,
    mime_type   TEXT NOT NULL,           -- audio/mpeg ou audio/wav
    size_bytes  BIGINT NOT NULL,
    bitrate_kbps INTEGER,
    object_key  TEXT NOT NULL UNIQUE,    -- caminho no MinIO
    checksum    TEXT NOT NULL,           -- SHA-256
    created_at  TIMESTAMPTZ DEFAULT now(),
    updated_at  TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_tracks_search ON tracks
  USING GIN (to_tsvector('simple', title || ' ' || artist));
```

### 3. Posições de Reprodução (Redis)
As posições são escritas na aplicação frequentemente (a cada avanço de reprodução). Elas também exigem baixa latência e tempo de vida curto. Por isso, para cumprir os requisitos, foi escolhido o **Redis** como forma de armazenamento.

O formato da chave string é `playback:<user_id>:<track_id>`. Já o valor tem o formato `{"posicao_seg": 123.4, "atualizado_em": "..."}`, com TTL configurável.

## Estratégia do Buffer
A reprodução de áudio no cliente é feita pelo elemento <audio> nativo do navegador, apoiado pelo protocolo HTTP definido no backend. Como o navegador já implementa, de forma nativa e otimizada, o controle de buffer, cabe ao frontend apenas observar e reagir aos eventos do <audio>.

### Tamanho dos Chunks
O navegador decide, por conta própria, quantos bytes solicitar a cada requisição. Para isso, ele faz requisições Range ao backend. O backend responde 206 Partial Content com o trecho solicitado.

### Nível Mínimo para Reprodução
O início da reprodução é definido pelos eventos do <audio> ao invés de um número fixo de segundos. O que ocorre é que o navegador tenta reproduzir o áudio logo depois de receber uma quantidade suficiente de dados pra começar a decodificar.

### Nível Máximo para Reprodução
O navegador, de forma nativa, para automaticamente de baixar quando o buffer do <audio> fica cheio. Logo, a determinação de um teto é opcional. Pode ser definido um teto de 30 a 40s para calibrar o desempenho da reprodução.

### Estratégia contra Jitter
O <audio> absorve o jitter por meio do buffer à frente do ponto de reprodução (bufferedAhead). Com isso, o papel do frontend é acobertar visualmente as oscilações de rede.
Na prática, enquanto houver dados no buffer à frente, a reprodução irá continuar msmo que a rede oscile. Caso isso aconteça, o <audio> dispara o evento waiting e pausa a reprodução. O client então vai, por exemplo, exibir um spinner de carregamento até a oscilação passar.

### Pausar e Avançar na Faixa
Para pausar e retomar, foi escolhido usar os métodos nativos `.pause()` e `.play()` do elemento <audio>.
Já para avançar, foi definido que o frontend irá alterar o valor de `.currentTime` para `posicao_atual` e o navegador vai disparar uma requisição Range correspondente ao backend. O backend responde 206 com o trecho e a reprodução retoma do novo ponto. Com isso, o frontend não precisa construir a requisição Range manualmente.

## Protocolo de Transmissão
O protocolo de transmissão escolhido é HTTP/1.1 sobre REST, com suporte a Range requests

### Rotas:
```text
GET api/tracks               # Catálogo
GET api/tracks/{id}          # Metadados
GET api/tracks/{id}/stream   # Áudio
GET /api/playback/{trackId}  # Última posição do usuário
PUT /api/playback/{trackId}  # Salva posição de reprodução
GET /healthz                 # Health Check
```
**OBS:** `{id}` e `{trackId}` são UUIDs v4.

### Rota de Catálogo
Request:
```text
GET /api/tracks HTTP/1.1
Host: <ip-virtual-do-load-balancer>
Accept: application/json
```

#### Response 200:
HTTP:
```text
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Cache-Control: public, max-age=60
```
JSON:
```text
{
  "total": 1,
  "itens": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "titulo": "audio1",
      "formato": "mp3",
      "mime_type": "audio/mpeg",
      "duracao_seg": 360,
      "tamanho_bytes": 5760000,
      "stream_url": "/api/tracks/550e8400-e29b-41d4-a716-446655440000/stream"
    }
  ]
}
```

#### Response 503 (DB indisponível):
HTTP:
```text
HTTP/1.1 503 Service Unavailable
Content-Type: application/json; charset=utf-8
Cache-Control: no-store
```
JSON:
```text
{"erro":"db_indisponivel"}
```

### Rota de Metadados
Request:
```text
GET /api/tracks/550e8400-e29b-41d4-a716-446655440000 HTTP/1.1
Host: <ip-virtual-do-load-balancer>
Accept: application/json
```

#### Response 200:
HTTP:
```text
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Cache-Control: public, max-age=60
```
JSON:
```text
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "titulo": "audio1",
  "formato": "mp3",
  "mime_type": "audio/mpeg",
  "duracao_seg": 360,
  "tamanho_bytes": 5760000,
  "stream_url": "/api/tracks/550e8400-e29b-41d4-a716-446655440000/stream"
}
```

#### Response 404 (Track não encontrada):
HTTP:
```text
HTTP/1.1 404 Not Found
Content-Type: application/json; charset=utf-8
```
JSON:
```text
{"erro":"track_nao_encontrada"}
```

### Rota de Stream
#### Sem Range:
Request:
```text
GET /api/tracks/550e8400-e29b-41d4-a716-446655440000/stream HTTP/1.1
Host: <ip-virtual-do-load-balancer>
```

HTTP Response 200:
```text
HTTP/1.1 200 OK
Content-Type: audio/mpeg
Content-Length: 5760000
Accept-Ranges: bytes
Cache-Control: public, max-age=3600
```

#### Com Range:
Request satisfazível:
```text
GET /api/tracks/550e8400-e29b-41d4-a716-446655440000/stream HTTP/1.1
Host: <ip-virtual-do-load-balancer>
Range: bytes=1048576-2097151
```

HTTP Response 206:
```text
HTTP/1.1 206 Partial Content
Content-Type: audio/mpeg
Content-Range: bytes 1048576-2097151/5760000
Content-Length: 1048576
Accept-Ranges: bytes
Cache-Control: public, max-age=3600
```

Request não satisfazível:
```text
GET /api/tracks/550e8400-e29b-41d4-a716-446655440000/stream HTTP/1.1
Range: bytes=99999999-100000000
```

HTTP Response 416:
```text
HTTP/1.1 416 Range Not Satisfiable
Content-Type: text/plain; charset=utf-8
Content-Range: bytes */5760000
Content-Length: 0
```

### Rota de Última Posição de Reprodução
Request:
```text
GET /api/playback/550e8400-e29b-41d4-a716-446655440000 HTTP/1.1
Accept: application/json
```

#### Response 200:
HTTP:
```text
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Cache-Control: private, no-store
```
JSON:
```text
{
  "track_id": "550e8400-e29b-41d4-a716-446655440000",
  "posicao_seg": 123.4,
  "atualizado_em": "2026-11-15T14:32:00Z"
}
```

#### Response 404 (Nenhuma posição encontrada):
HTTP:
```text
HTTP/1.1 404 Not Found
Content-Type: application/json; charset=utf-8
```
JSON:
```text
{"erro":"posicao_nao_encontrada"}
```

### Rota de Salvar Posição
Request HTTP:
```text
PUT /api/playback/550e8400-e29b-41d4-a716-446655440000 HTTP/1.1
Content-Type: application/json; charset=utf-8
```
Request JSON:
```text
{"posicao_seg": 123.4}
```

#### Response 204:
HTTP:
```text
HTTP/1.1 204 No Content
Cache-Control: no-store
```

#### Response 422 (Payload inválido):
HTTP:
```text
HTTP/1.1 422 Unprocessable Entity
Content-Type: application/json; charset=utf-8
```
JSON:
```text
{"erro":"posicao_invalida"}
```

#### Response 404 (Track inexistente):
JSON:
```text
{"erro":"track_nao_encontrada"}
```

### Rota de Health Check
Request HTTP:
```text
GET /healthz HTTP/1.1
```
Request JSON:
```text
{"status":"ok"}
```

#### Response 200:
HTTP:
```text
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Cache-Control: no-store
```

#### Response 503 (Serviço indisponível):
HTTP:
```text
HTTP/1.1 503 Service Unavailable
Content-Type: application/json; charset=utf-8
Cache-Control: no-store
```
JSON:
```text
{"status":"degraded","dependencias":{"db":"indisponivel"}}
```

### Autores:
- [@adnalisia](https://github.com/adnalisia)
- [@brunoslm](https://github.com/brunoslm)
- [@Pedro-glitch-370](https://github.com/Pedro-glitch-370)
- [@Valci2](https://github.com/Valci2)
- [@Vinicius2703](https://github.com/Vinicius2703)