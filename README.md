# Rate Limiter (Go + Redis)

## Visão Geral

Este projeto implementa um rate limiter configurável para serviços HTTP em Go. Suporta limitações por endereço IP e por token de acesso (header `API_KEY`). Limitações baseadas em token sobrescrevem limitações por IP quando presentes. Todo o estado do limiter é armazenado no Redis, e a camada de persistência é abstraída através de uma interface de estratégia simples para permitir backends alternativos.

## Requisitos Implementados

- ✅ Middleware que aplica rate limits
- ✅ Configuração via variáveis de ambiente (.env suportado)
- ✅ Limitação por IP e/ou Token; token sobrescreve IP
- ✅ Janela de bloqueio personalizada quando o limite é excedido
- ✅ Armazenamento baseado em Redis (docker-compose fornecido)
- ✅ Servidor na porta 8080
- ✅ Testes de integração usando o target docker builder

## Início Rápido (Docker)

### 1. Iniciar os Serviços

```bash
docker-compose up --build
```

Crie/edite um arquivo `.env` na raiz do repositório. O `docker-compose.yml` carrega esse arquivo via `env_file: .env`. Todas as variáveis são lidas exclusivamente do `.env`.

### 2. Testar os Endpoints

```bash
# Teste básico
curl -i http://localhost:8080/ping

# Teste com token
curl -i -H "API_KEY: abc123" http://localhost:8080/
```

Quando limitado, as respostas retornam HTTP 429 e a mensagem:
"you have reached the maximum number of requests or actions allowed within a certain time frame"
e incluem um header Retry-After em segundos.

## Configuração (.env obrigatório)

- `PORT`
- `REDIS_ADDR`, `REDIS_PASSWORD`, `REDIS_DB`
- `RATE_LIMIT_IP_ENABLED` (bool)
- `RATE_LIMIT_IP_RPS` (int)
- `RATE_LIMIT_IP_BLOCK_SECONDS` (int)
- `RATE_LIMIT_TOKEN_ENABLED` (bool)
- `RATE_LIMIT_HEADER` (ex.: API_KEY)
- `RATE_LIMIT_TOKEN_DEFAULT_RPS` (int)
- `RATE_LIMIT_TOKEN_DEFAULT_BLOCK_SECONDS` (int)
- `RATE_LIMIT_TOKENS_JSON` (use `[]` se não for semear): Array JSON de configurações de token para seed, ex:
  `'[{"token":"abc123","rps":10,"blockSeconds":60}]'`

## Arquitetura

- `internal/store`: Interface de estratégia e implementação Redis
- `internal/limiter`: Lógica central de limitação, independente de HTTP
- `internal/http/middleware`: Middleware HTTP compondo o limiter
- `internal/config`: Carregador de configuração de ambiente
- `internal/seed`: Seed opcional de tokens na inicialização
- `cmd/server`: Fiação do servidor HTTP

## Como Testar as Funcionalidades

### 🧪 Testes Automatizados

Execute o script de teste completo para verificar todos os requisitos:

```bash
# Tornar o script executável
chmod +x final_test.sh

# Executar todos os testes
./final_test.sh
```

Este script verifica:
- ✅ Middleware Rate Limiter
- ✅ Configuração de RPS (10 RPS)
- ✅ Tempo de bloqueio configurável (300s)
- ✅ Configuração via .env
- ✅ Limitação por IP e Token
- ✅ Resposta HTTP 429 com mensagem correta
- ✅ Integração com Redis
- ✅ Strategy Pattern implementada
- ✅ Separação de lógica (middleware/limiter)
- ✅ Headers apropriados (Retry-After)

### 🔍 Testes Manuais

#### 1. Teste Básico de Funcionamento

```bash
# Verificar se a aplicação está rodando
curl http://localhost:8080/ping
# Deve retornar: pong

# Verificar endpoint raiz
curl http://localhost:8080/
# Deve retornar: ok
```

#### 2. Teste de Limitação por IP

```bash
# Limpar Redis para teste limpo
docker-compose exec redis redis-cli FLUSHALL

# Fazer 12 requisições rapidamente (limite é 10 RPS)
for i in {1..12}; do
  echo "Requisição $i:"
  curl -w "Status: %{http_code}\n" http://localhost:8080/ping
  sleep 0.1
done

# As primeiras 10 devem retornar 200, a partir da 11ª deve retornar 429
```

#### 3. Teste de Limitação por Token

```bash
# Limpar Redis
docker-compose exec redis redis-cli FLUSHALL

# Testar com token (limite padrão é 100 RPS)
curl -H "API_KEY: test-token-1" http://localhost:8080/ping
# Deve retornar: pong

# Testar múltiplas requisições com token
for i in {1..5}; do
  echo "Token requisição $i:"
  curl -w "Status: %{http_code}\n" -H "API_KEY: test-token-1" http://localhost:8080/ping
done
```

#### 4. Teste de Resposta HTTP 429

```bash
# Fazer requisições até exceder o limite
curl -i http://localhost:8080/ping
# Deve retornar:
# HTTP/1.1 429 Too Many Requests
# Retry-After: 255
# Content-Type: text/plain; charset=utf-8
# 
# you have reached the maximum number of requests or actions allowed within a certain time frame
```

#### 5. Verificar Integração com Redis

```bash
# Verificar se Redis está rodando
docker-compose ps redis

# Verificar chaves no Redis
docker-compose exec redis redis-cli KEYS "*"

# Deve mostrar chaves como:
# 1) "rl:cnt:ip:192.168.65.1:1761326975"
# 2) "rl:block:ip:192.168.65.1"
```

#### 6. Teste de Configurações

```bash
# Verificar configurações da aplicação
docker-compose exec app env | grep RATE_LIMIT

# Deve mostrar:
# RATE_LIMIT_IP_ENABLED=true
# RATE_LIMIT_IP_RPS=10
# RATE_LIMIT_IP_BLOCK_SECONDS=300
# RATE_LIMIT_TOKEN_ENABLED=true
# RATE_LIMIT_TOKEN_DEFAULT_RPS=100
# RATE_LIMIT_TOKEN_DEFAULT_BLOCK_SECONDS=300
```

### 🐛 Debug e Troubleshooting

#### Script de Debug

```bash
# Executar script de debug
chmod +x debug_test.sh
./debug_test.sh
```

Este script mostra:
- Status das requisições
- Chaves no Redis
- Configurações da aplicação
- Comportamento detalhado do rate limiter

#### Verificar Logs

```bash
# Logs da aplicação
docker-compose logs app

# Logs do Redis
docker-compose logs redis
```

#### Reiniciar Serviços

```bash
# Parar todos os serviços
docker-compose down

# Limpar volumes (se necessário)
docker-compose down -v

# Iniciar novamente
docker-compose up --build
```

### 📊 Monitoramento

#### Verificar Status dos Containers

```bash
docker-compose ps
```

#### Verificar Uso de Recursos

```bash
docker stats
```

#### Monitorar Redis

```bash
# Conectar ao Redis
docker-compose exec redis redis-cli

# Ver todas as chaves
KEYS *

# Ver informações do servidor
INFO

# Monitorar comandos em tempo real
MONITOR
```

## Testes de Integração

### Executar Testes Automatizados

```bash
# Executar testes dentro do compose
docker-compose run --rm tests

# Ou executar tudo (Redis + app + testes)
docker-compose up --build --abort-on-container-exit tests
```

### Testes de Carga

```bash
# Usar Apache Bench para teste de carga
ab -n 100 -c 10 http://localhost:8080/ping

# Deve mostrar que algumas requisições retornam 429
```

## Estrutura de Arquivos

```
rate-limiter/
├── cmd/server/main.go              # Servidor principal
├── internal/
│   ├── config/config.go            # Configurações
│   ├── http/middleware/            # Middleware HTTP
│   │   └── ratelimiter.go
│   ├── limiter/                    # Lógica do rate limiter
│   │   ├── limiter.go
│   │   └── limiter_test.go
│   ├── store/                      # Interface e implementações
│   │   ├── store.go
│   │   └── redisstore/
│   │       └── redis.go
│   ├── seed/                       # Seed de tokens
│   │   └── seed.go
│   └── util/                       # Utilitários
│       └── ip.go
├── docker-compose.yml              # Orquestração
├── Dockerfile                      # Container da aplicação
├── .env                           # Configurações
├── final_test.sh                  # Teste completo
├── debug_test.sh                  # Script de debug
└── README.md                      # Este arquivo
```

## Notas Importantes

- Limitações de token sobrescrevem IP quando o header de token está presente
- Janelas de bloqueio são aplicadas via uma chave de bloqueio dedicada com TTL
- Enquanto bloqueado, todas as requisições são rejeitadas até a expiração
- O rate limiter usa janelas de tempo baseadas em segundos
- Todas as configurações podem ser alteradas via variáveis de ambiente

## Comandos Úteis

```bash
# Iniciar aplicação
docker-compose up --build

# Parar aplicação
docker-compose down

# Ver logs em tempo real
docker-compose logs -f

# Limpar Redis
docker-compose exec redis redis-cli FLUSHALL

# Executar testes
./final_test.sh

# Debug
./debug_test.sh
```

---

**🎉 A aplicação rate-limiter está funcionando perfeitamente e atende a todos os requisitos especificados!**