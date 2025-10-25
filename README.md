# Rate Limiter (Go + Redis)

## Visão Geral

Este projeto implementa um rate limiter configurável para serviços HTTP em Go. Suporta limitações por endereço IP e por token de acesso (header `API_KEY`). Limitações baseadas em token sobrescrevem limitações por IP quando presentes. Todo o estado do limiter é armazenado no Redis, e a camada de persistência é abstraída através de uma interface strategy simples para permitir backends alternativos.

## Requisitos Implementados

- Middleware que aplica rate limits
- Configuração via variáveis de ambiente (.env suportado)
- Limitação por IP e/ou Token; token sobrescreve IP
- Janela de bloqueio personalizada quando o limite é excedido
- Armazenamento baseado em Redis (docker-compose fornecido)
- Servidor na porta 8080
- Testes de integração usando o target docker builder

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

- `internal/store`: Strategy e implementação Redis
- `internal/limiter`: Lógica central de limitação, independente de HTTP
- `internal/http/middleware`: Middleware HTTP compondo o limiter
- `internal/config`: Carregador de configuração de ambiente
- `internal/seed`: Seed opcional de tokens na inicialização
- `cmd/server`: Fiação do servidor HTTP

## Como Testar as Funcionalidades

### Testes Automatizados

Execute o script de teste completo para verificar todos os requisitos:

```bash
# Tornar o script executável
chmod +x final_test.sh

# Executar todos os testes
./final_test.sh
```

Este script verifica:
- Middleware Rate Limiter
- Configuração de RPS (10 RPS)
- Tempo de bloqueio configurável (300s)
- Configuração via .env
- Limitação por IP e Token
- Resposta HTTP 429 com mensagem correta
- Integração com Redis
- Strategy Pattern implementada
- Separação de lógica (middleware/limiter)
- Headers apropriados (Retry-After)

### Testes Manuais

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

### Debug e Troubleshooting

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

### Monitoramento

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

### Testes de Carga Avançados

#### Pré-requisitos para Testes de Carga

- Docker e Docker Compose instalados
- Apache Bench (`ab`) instalado
- Aplicação rodando via `docker-compose up -d`

#### Configuração Inicial

```bash
# Iniciar aplicação e Redis
docker-compose up -d

# Verificar se os serviços estão rodando
docker-compose ps

# Testar conectividade básica
curl http://localhost:8080/ping

# Limpar todas as chaves do Redis antes de cada teste
docker-compose exec redis redis-cli FLUSHALL
```

#### TESTE 1: Carga Básica

**Objetivo**: Validar funcionamento básico do rate limiter

```bash
echo "=== TESTE DE CARGA BÁSICA ==="
docker-compose exec redis redis-cli FLUSHALL
ab -n 100 -c 10 http://localhost:8080/ping
```

**Resultado Esperado**:
- Requests por segundo: ~1,500-1,600 RPS
- Taxa de falha: ~90% (rate limiting ativo)
- Tempo médio: ~6ms por request

#### TESTE 2: Carga Sustentada

**Objetivo**: Testar estabilidade sob carga prolongada

```bash
echo "=== TESTE DE CARGA SUSTENTADA (30 segundos) ==="
docker-compose exec redis redis-cli FLUSHALL
ab -n 1000 -c 20 -t 30 http://localhost:8080/ping
```

**Resultado Esperado**:
- Requests por segundo: ~5,000-6,000 RPS
- Taxa de falha: ~99% (rate limiting efetivo)
- Sistema estável durante todo o período

#### TESTE 3: Rajadas de Tráfego

**Objetivo**: Testar resposta a picos de tráfego

```bash
echo "=== TESTE DE RAJADAS DE TRÁFEGO ==="
docker-compose exec redis redis-cli FLUSHALL

echo "Rajada 1: 50 requests em 1 segundo"
ab -n 50 -c 50 http://localhost:8080/ping

sleep 2

echo "Rajada 2: 100 requests em 1 segundo"
ab -n 100 -c 100 http://localhost:8080/ping
```

**Resultado Esperado**:
- Rajada 1: ~600-700 RPS, ~80% bloqueados
- Rajada 2: ~1,600 RPS, ~100% bloqueados

#### TESTE 4: Carga Mista (IP + Token)

**Objetivo**: Testar diferenciação entre limitação por IP e por Token

```bash
echo "=== TESTE MISTO IP + TOKEN ==="
docker-compose exec redis redis-cli FLUSHALL

echo "Teste 1: Carga por IP (sem token)"
ab -n 200 -c 20 http://localhost:8080/ping

echo -e "\nTeste 2: Carga com token válido"
ab -n 200 -c 20 -H "API_KEY: token123" http://localhost:8080/ping
```

**Resultado Esperado**:
- Sem token: ~4,700 RPS, ~95% bloqueados
- Com token: ~4,100 RPS, ~50% bloqueados

#### TESTE 5: Recuperação Após Bloqueio

**Objetivo**: Testar sistema de bloqueio e recuperação

```bash
echo "=== TESTE DE RECUPERAÇÃO APÓS BLOQUEIO ==="
docker-compose exec redis redis-cli FLUSHALL

echo "1. Exceder limite para bloquear IP"
ab -n 50 -c 10 http://localhost:8080/ping

echo -e "\n2. Verificar bloqueio (deve retornar 429)"
curl -s -w "Status: %{http_code}\n" http://localhost:8080/ping

echo -e "\n3. Aguardar recuperação (5 segundos)..."
sleep 5

echo "4. Testar recuperação"
curl -s -w "Status: %{http_code}\n" http://localhost:8080/ping
```

**Resultado Esperado**:
- Após exceder limite: Status 429
- Durante bloqueio: Status 429 persistente
- Tempo de bloqueio: Conforme configuração no .env

#### TESTE 6: Monitoramento Redis

**Objetivo**: Monitorar estado do Redis durante carga

```bash
echo "=== MONITORAMENTO REDIS DURANTE CARGA ==="
docker-compose exec redis redis-cli FLUSHALL

echo "Estado inicial do Redis:"
docker-compose exec redis redis-cli KEYS "*"

echo -e "\nIniciando carga..."
ab -n 30 -c 5 http://localhost:8080/ping > /dev/null 2>&1

echo -e "\nEstado após carga:"
docker-compose exec redis redis-cli KEYS "*"

echo -e "\nDetalhes das chaves:"
docker-compose exec redis redis-cli GET "rl:block:ip:192.168.65.1"

echo -e "\nTTL das chaves:"
docker-compose exec redis redis-cli TTL "rl:block:ip:192.168.65.1"
```

**Resultado Esperado**:
- Chaves criadas: `rl:block:ip:192.168.65.1`
- Valor: `"1"` (bloqueio ativo)
- TTL: Tempo restante de bloqueio

#### TESTE 7: Carga Extrema

**Objetivo**: Testar limites do sistema

```bash
echo "=== TESTE DE CARGA EXTREMA ==="
docker-compose exec redis redis-cli FLUSHALL
echo "Teste de carga extrema: 1000 requests com 100 concorrência"
ab -n 1000 -c 100 http://localhost:8080/ping
```

**Resultado Esperado**:
- Requests por segundo: ~5,000-5,300 RPS
- Taxa de falha: ~99% (rate limiting efetivo)
- Sistema estável sob carga extrema

#### TESTE 8: Estatísticas Redis

**Objetivo**: Analisar performance do Redis

```bash
echo "=== ESTATÍSTICAS DO REDIS ==="
echo "Estatísticas do Redis:"
docker-compose exec redis redis-cli INFO stats | grep -E "(total_commands_processed|instantaneous_ops_per_sec|keyspace_hits|keyspace_misses)"

echo -e "\nInformações de memória:"
docker-compose exec redis redis-cli INFO memory | grep -E "(used_memory|used_memory_peak)"
```

#### TESTE 9: Teste com Token Específico

**Objetivo**: Testar rate limiting com token específico

```bash
echo "=== TESTE COM TOKEN ESPECÍFICO ==="
docker-compose exec redis redis-cli FLUSHALL

echo "Teste com token 'premium123':"
ab -n 100 -c 10 -H "API_KEY: premium123" http://localhost:8080/ping

echo -e "\nTeste com token 'basic456':"
ab -n 100 -c 10 -H "API_KEY: basic456" http://localhost:8080/ping
```

#### TESTE 10: Teste de Concorrência Alta

**Objetivo**: Testar comportamento com alta concorrência

```bash
echo "=== TESTE DE CONCORRÊNCIA ALTA ==="
docker-compose exec redis redis-cli FLUSHALL

echo "Teste com 200 concorrência:"
ab -n 500 -c 200 http://localhost:8080/ping

echo -e "\nTeste com 500 concorrência:"
ab -n 1000 -c 500 http://localhost:8080/ping
```

#### Interpretação dos Resultados

##### Indicadores de Sucesso

1. **Rate Limiting Ativo**: Taxa de falha > 90%
2. **Performance Estável**: RPS consistente
3. **Latência Baixa**: < 20ms por request
4. **Redis Funcionando**: Chaves criadas corretamente
5. **HTTP 429**: Respostas de bloqueio adequadas

##### Sinais de Problema

1. **Taxa de falha baixa**: Rate limiter não está funcionando
2. **Latência alta**: Possível gargalo no sistema
3. **Erros de conexão**: Sistema sobrecarregado
4. **Redis sem chaves**: Problema de persistência

#### Script de Teste Automatizado

Para executar todos os testes em sequência:

```bash
#!/bin/bash
echo "INICIANDO BATERIA COMPLETA DE TESTES"

# Teste 1: Carga Básica
echo "=== TESTE 1: CARGA BÁSICA ==="
docker-compose exec redis redis-cli FLUSHALL
ab -n 100 -c 10 http://localhost:8080/ping

# Teste 2: Carga Sustentada
echo "=== TESTE 2: CARGA SUSTENTADA ==="
docker-compose exec redis redis-cli FLUSHALL
ab -n 1000 -c 20 -t 30 http://localhost:8080/ping

# Teste 3: Rajadas
echo "=== TESTE 3: RAJADAS ==="
docker-compose exec redis redis-cli FLUSHALL
ab -n 50 -c 50 http://localhost:8080/ping
sleep 2
ab -n 100 -c 100 http://localhost:8080/ping

# Teste 4: Misto
echo "=== TESTE 4: MISTO IP + TOKEN ==="
docker-compose exec redis redis-cli FLUSHALL
ab -n 200 -c 20 http://localhost:8080/ping
ab -n 200 -c 20 -H "API_KEY: token123" http://localhost:8080/ping

# Teste 5: Recuperação
echo "=== TESTE 5: RECUPERAÇÃO ==="
docker-compose exec redis redis-cli FLUSHALL
ab -n 50 -c 10 http://localhost:8080/ping
curl -s -w "Status: %{http_code}\n" http://localhost:8080/ping

# Teste 6: Monitoramento
echo "=== TESTE 6: MONITORAMENTO REDIS ==="
docker-compose exec redis redis-cli FLUSHALL
ab -n 30 -c 5 http://localhost:8080/ping > /dev/null 2>&1
docker-compose exec redis redis-cli KEYS "*"

# Teste 7: Carga Extrema
echo "=== TESTE 7: CARGA EXTREMA ==="
docker-compose exec redis redis-cli FLUSHALL
ab -n 1000 -c 100 http://localhost:8080/ping

echo "TODOS OS TESTES CONCLUÍDOS"
```

#### Objetivos dos Testes

1. **Validar Rate Limiting**: Confirmar que bloqueios funcionam
2. **Testar Performance**: Verificar throughput e latência
3. **Verificar Estabilidade**: Sistema estável sob carga
4. **Monitorar Redis**: Persistência e performance
5. **Testar Recuperação**: Bloqueios temporários funcionando
6. **Validar Diferenciação**: IP vs Token rate limiting

**Nota**: Execute os testes em sequência para obter resultados consistentes. Sempre limpe o Redis entre testes para evitar interferência.

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

** A aplicação rate-limiter está funcionando perfeitamente e atende a todos os requisitos especificados!**