# Airline Booking API - Sistema de Reserva de Assentos (Passagens Aéreas)

API completa em Go que implementa um sistema de reserva inteligente com bloqueio de assento para agência de passagens aéreas, usando MySQL (transacional) e Elasticsearch (busca/consulta).

## 🎯 Características Principais

- **Bloqueio de Assento com TTL**: Assentos ficam reservados por 15 minutos (configurável)
- **Race-Condition Safe**: Uso de Compare-And-Set (CAS) para evitar condições de corrida
- **Idempotência**: Suporte a `Idempotency-Key` para operações críticas
- **Busca Avançada**: Elasticsearch para busca rápida e filtros
- **Rate Limiting**: Proteção contra spam
- **Auto-Cleanup**: Job automático para limpeza de holds expirados

## 🛠️ Stack Tecnológica

- **Linguagem**: Go 1.23+
- **Framework HTTP**: Gin
- **Banco Transacional**: MySQL 8+
- **Busca**: Elasticsearch 8+
- **Indexação Automática**: Sincronização MySQL ↔ Elasticsearch
- **ORM/Query**: SQLC
- **Migrations**: golang-migrate
- **Logs**: Zap (estruturados)
- **Config**: 12-factor via env
- **Container**: Docker + docker-compose

## 🚀 Quick Start

### Pré-requisitos

- Docker e Docker Compose
- Make (recomendado para instalação automatizada)
- Go 1.23+ (para desenvolvimento local)

### 1. Clone e Configure

```bash
git clone <repository>
cd airline-booking-api

# Copie e ajuste as variáveis de ambiente se necessário
cp .env.example .env
```

### 2. Instalação Completa Automatizada (Recomendado)

```bash
# 🚀 Instala TUDO do zero: builds, migrations, seeds e sincronização ES
make install
```

Este comando único executa:
- Para e remove containers existentes
- Rebuilda a aplicação
- Inicia todos os serviços (MySQL, Elasticsearch, Kibana, API)
- Executa migrations automaticamente
- Popula o banco com dados de demonstração
- Sincroniza dados com Elasticsearch
- Configura todos os índices necessários

### 3. Método Manual (Alternativo)

```bash
# Se preferir executar passo a passo:
make up              # Sobe os serviços
make migrate-up      # Executa migrations
make seed           # Popula dados de teste
make es-seed        # Sincroniza com Elasticsearch
```

### 4. Teste a API

A API estará disponível em `http://localhost:8080`

```bash
# Health check
curl http://localhost:8080/api/v1/health

# Buscar voos
curl "http://localhost:8080/api/v1/flights/search?origin=JFK&destination=LAX&date=2025-08-30"

# Ver assentos disponíveis
curl http://localhost:8080/api/v1/flights/1/seats
```

## ✅ Sistema Completamente Funcional

O sistema implementa as seguintes funcionalidades:

- ✅ **API de Criação de Flights** com indexação automática no Elasticsearch
- ✅ **Sistema de Hold/Reserve** com TTL de 15 minutos
- ✅ **Confirmação de Tickets** com criação de PNR automático
- ✅ **Sincronização Automática** entre MySQL e Elasticsearch
- ✅ **Indexação de 3 Entidades**: flights, holds e tickets
- ✅ **Fluxo Completo Funcional**: Flight → Hold → Ticket → ES Sync

## 📋 Endpoints da API

### Busca de Voos
```
GET /api/v1/flights/search
```
**Query Params**: `origin`, `destination`, `date`, `fare_class?`, `airline?`, `page?`, `size?`

### Disponibilidade de Assentos
```
GET /api/v1/flights/{id}/seats
```

### Criar Hold (Bloqueio)
```
POST /api/v1/holds
Headers: User-ID, Idempotency-Key?
Body: {"flight_id": 1, "seat_no": "12A"}
```

### Liberar Hold
```
DELETE /api/v1/holds/{flight_id}/{seat_no}
Headers: User-ID
```

### Confirmar Compra
```
POST /api/v1/tickets/confirm
Headers: User-ID, Idempotency-Key?
Body: {"flight_id": 1, "seat_no": "12A", "payment_ref": "pay_123"}
```

## 📊 Dados de Demonstração

O projeto inclui um conjunto abrangente de dados de demonstração que é automaticamente carregado:

### ✈️ Voos Incluídos
- **15 voos** cobrindo rotas domésticas e internacionais
- **JFK ↔ LAX**: Voos da American Airlines e Delta
- **Rotas Internacionais**: JFK-LHR, LAX-NRT, MIA-GRU
- **Diferentes Classes**: Economy, Business e First Class
- **Múltiplas Companhias**: AA, DL, UA, SW, BA, VS, JL, ANA

### � Assentos
- **Voo 1 (JFK-LAX)**: 162 assentos (First: 12, Business: 20, Economy: 130)
- **Voo 2 (JFK-LAX)**: 142 assentos (Business: 16, Economy: 126)
- **Outros voos**: Configurações variadas para teste

### 📋 Status dos Assentos
- **Disponíveis**: Maioria dos assentos
- **Em Hold**: 4 assentos com diferentes tempos de expiração
- **Vendidos**: 9 tickets confirmados com PNRs

### 💳 Tickets Confirmados
- **PNR Codes**: ABC001, ABC002, ABC003, etc.
- **Preços Realistas**: $299 (Economy) a $1,499 (First Class)
- **Diferentes Usuários**: customer001 a customer009

### 🔄 Como Recarregar os Dados

```bash
# Método 1: Reinstalação completa (recomendado)
make install

# Método 2: Usar o seeder Go apenas
make seed

# Método 3: Usar arquivo SQL direto  
make seed-sql

```

## 🧪 Exemplos de Uso com Dados Reais

### Fluxo Completo de Reserva

```bash
# 1. Buscar voos JFK → LAX
curl -G "http://localhost:8080/api/v1/flights/search" \
  -d "origin=JFK" \
  -d "destination=LAX" \
  -d "date=2025-08-30"

# 2. Ver assentos disponíveis do voo 1
curl "http://localhost:8080/api/v1/flights/1/seats"

# 3. Criar hold no assento 25A (disponível)
curl -X POST "http://localhost:8080/api/v1/holds" \
  -H "Content-Type: application/json" \
  -H "User-ID: testuser123" \
  -H "Idempotency-Key: $(uuidgen)" \
  -d '{"flight_id": 1, "seat_no": "25A"}'

# 4. Confirmar compra (dentro de 15 min)
curl -X POST "http://localhost:8080/api/v1/tickets/confirm" \
  -H "Content-Type: application/json" \
  -H "User-ID: testuser123" \
  -H "Idempotency-Key: $(uuidgen)" \
  -d '{"flight_id": 1, "seat_no": "25A", "payment_ref": "payment_789"}'
```

### Testar Cenários Específicos

```bash
# Tentar hold em assento já vendido (10A) - deve retornar 409
curl -X POST "http://localhost:8080/api/v1/holds" \
  -H "Content-Type: application/json" \
  -H "User-ID: testuser456" \
  -d '{"flight_id": 1, "seat_no": "10A"}'

# Tentar hold em assento já em hold (12A) - deve retornar 409
curl -X POST "http://localhost:8080/api/v1/holds" \
  -H "Content-Type: application/json" \
  -H "User-ID: testuser789" \
  -d '{"flight_id": 1, "seat_no": "12A"}'

# Buscar voos internacionais
curl -G "http://localhost:8080/api/v1/flights/search" \
  -d "origin=JFK" \
  -d "destination=LHR" \
  -d "date=2025-08-30"
```

### Teste de Concorrência

```bash
# Teste: múltiplos usuários tentando o mesmo assento
for i in {1..5}; do
  curl -X POST "http://localhost:8080/api/v1/holds" \
    -H "Content-Type: application/json" \
    -H "User-ID: user$i" \
    -d '{"flight_id": 1, "seat_no": "15B"}' &
done
wait
# Apenas um deve ser bem-sucedido (201), outros receberão 409 Conflict
```

### Teste de Expiração

```bash
# 1. Configurar TTL curto (1 minuto) no .env
echo "HOLD_TTL_MINUTES=1" >> .env

# 2. Reiniciar serviço
make down && make up

# 3. Criar hold
curl -X POST "http://localhost:8080/api/v1/holds" \
  -H "Content-Type: application/json" \
  -H "User-ID: user123" \
  -d '{"flight_id": 1, "seat_no": "20A"}'

# 4. Aguardar 1+ minuto, depois tentar novamente com outro usuário
sleep 70
curl -X POST "http://localhost:8080/api/v1/holds" \
  -H "Content-Type: application/json" \
  -H "User-ID: user456" \
  -d '{"flight_id": 1, "seat_no": "20A"}'
# Deve ser bem-sucedido após expiração
```

## 🏗️ Arquitetura

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Gin Router    │    │  Booking Service │    │   Repositories  │
│                 │────│                  │────│                 │
│ • Rate Limiting │    │ • Hold Logic     │    │ • Seat Repo     │
│ • Middleware    │    │ • Idempotency    │    │ • Ticket Repo   │
│ • CORS          │    │ • Validation     │    │ • Flight Repo   │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                               │
        ┌──────────────────────┼──────────────────────┐
        │                      │                      │
┌───────▼────────┐    ┌────────▼────────┐    ┌────────▼────────┐
│     MySQL      │    │ Elasticsearch   │    │   Cleanup Job   │
│                │    │                 │    │                 │
│ • Transactions │    │ • Flight Search │    │ • Expired Holds │
│ • ACID         │    │ • Aggregations  │    │ • Cron Schedule │
│ • CAS Logic    │    │ • Fast Queries  │    │ • Auto Cleanup  │
└────────────────┘    └─────────────────┘    └─────────────────┘
```

### Fluxo de Dados - Criação de Hold

```
1. Client POST /holds → Rate Limiter → Validation
2. Check Idempotency Key (if provided)
3. Validate Flight Exists (MySQL)
4. Check Existing Ticket (MySQL)
5. CAS Update/Insert Lock (MySQL) ← Race-condition safe
6. Store Idempotency Key (MySQL)
7. Return Response with Expiration
```

### Algoritmo de Bloqueio (CAS)

```sql
-- Tentativa 1: Inserir novo lock
INSERT INTO seat_locks (flight_id, seat_no, holder_id, expires_at) 
VALUES (?, ?, ?, ?) 
ON DUPLICATE KEY UPDATE ...

-- Tentativa 2: Se falhou, tentar atualizar com CAS
UPDATE seat_locks 
SET holder_id=?, expires_at=?, updated_at=NOW() 
WHERE flight_id=? AND seat_no=? 
AND (expires_at < NOW() OR holder_id=?)

-- Se rows_affected = 0 → Conflito (409)
-- Se rows_affected = 1 → Sucesso (201)
```

## 🧪 Testes

### Executar Testes

```bash
# Testes unitários
make test

# Testes com detecção de race conditions
make test-race

# Testes com cobertura
make test-cover
```

### Tipos de Teste

    1. **Unitários**: Lógica de negócio isolada
    2. **Concorrência**: N goroutines disputando mesmo assento
    3. **Integração**: Fluxo completo com banco real
    4. **E2E**: Testes via HTTP com docker-compose

## 📊 Monitoramento

### Logs Estruturados

```json
{
  "level": "info",
  "ts": "2025-08-29T10:30:00.000Z",
  "caller": "service/booking_service.go:45",
  "msg": "Hold created successfully",
  "flight_id": 1,
  "seat_no": "12A",
  "holder_id": "user123",
  "expires_at": "2025-08-29T10:45:00.000Z"
}
```

### Health Checks

```bash
curl http://localhost:8080/health
# Retorna status de saúde da aplicação
```

### Kibana (Opcional)

Acesse `http://localhost:5601` para visualizar logs e métricas do Elasticsearch.

## ⚙️ Configuração

### Variáveis de Ambiente (.env)

```bash
# Aplicação
APP_ENV=development
APP_PORT=8080
HOLD_TTL_MINUTES=15

# Banco de Dados
DB_HOST=localhost
DB_PORT=3306
DB_USER=airline_user
DB_PASSWORD=airline_pass
DB_NAME=airline_booking

# Elasticsearch
ES_ADDRESSES=http://localhost:9200

# Rate Limiting
RATE_LIMIT_PER_MINUTE=60

# Logs
LOG_LEVEL=info
LOG_FORMAT=json
```

### Configuração de Produção

```bash
# Para produção, ajuste:
APP_ENV=production
LOG_LEVEL=warn
HOLD_TTL_MINUTES=15
RATE_LIMIT_PER_MINUTE=100
```

## 🐛 Troubleshooting

### Problemas Comuns

1. **Erro de Conexão MySQL**
   ```bash
   make down && make up
   # Aguarde MySQL inicializar completamente
   ```

2. **Elasticsearch não responde**
   ```bash
   docker-compose logs elasticsearch
   # Verifique se tem memória suficiente
   ```

3. **Migrations falham**
   ```bash
   # Reset do banco
   make down
   docker volume rm $(docker volume ls -q | grep mysql)
   make up && make migrate-up
   ```

### Debug Mode

```bash
# Ativar logs debug
export LOG_LEVEL=debug
make dev
```

## 📈 Performance

### Benchmarks Esperados

- **Criação de Hold**: ~5ms (p95)
- **Busca de Voos**: ~20ms (p95)
- **Confirmação de Ticket**: ~10ms (p95)

### Otimizações Implementadas

1. **Índices MySQL**: `expires_at`, `(flight_id, seat_no)`
2. **Connection Pool**: 25 max, 5 idle
3. **ES Bulk Operations**: Para seed de dados
4. **Cleanup Job**: Roda a cada minuto (configurável)

## 🔒 Segurança

### Implementado

- ✅ Rate Limiting por IP
- ✅ Validação de payload
- ✅ SQL Injection safe (SQLC)
- ✅ CORS configurado
- ✅ Structured logging
- ✅ Idempotency keys

### TODO (Produção)

- [ ] HTTPS/TLS
- [ ] JWT Authentication
- [ ] Input sanitization
- [ ] Request timeout
- [ ] Circuit breaker

## 🚢 Deploy

### Docker Production

```bash
# Build da imagem
docker build -t airline-booking:latest .

# Run em produção
docker run -p 8080:8080 \
  -e APP_ENV=production \
  -e DB_HOST=prod-mysql \
  -e ES_ADDRESSES=http://prod-es:9200 \
  airline-booking:latest
```

### Kubernetes (Exemplo)

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: airline-booking
spec:
  replicas: 3
  selector:
    matchLabels:
      app: airline-booking
  template:
    metadata:
      labels:
        app: airline-booking
    spec:
      containers:
      - name: app
        image: airline-booking:latest
        ports:
        - containerPort: 8080
        env:
        - name: APP_ENV
          value: production
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
```

## 📚 Comandos Make

```bash
make install     # 🚀 INSTALAÇÃO COMPLETA: build + services + migrations + seeds + ES sync
make up          # Sobe todos os serviços
make down        # Para todos os serviços  
make migrate-up  # Executa migrations
make seed        # Popula banco de dados
make es-seed     # Popula Elasticsearch
make dev         # Executa em modo desenvolvimento
make test        # Executa testes
make test-race   # Testes com race detection
make build       # Build da aplicação
make clean       # Limpa artifacts
make urls        # Mostra URLs importantes após instalação
make help        # Lista todos os comandos
```

## 🎯 Casos de Uso Demonstrados

    1. **Concorrência**: ✅ Apenas um usuário consegue hold
    2. **Expiração**: ✅ Auto-liberação após TTL
    3. **Idempotência**: ✅ Mesma chave = mesma resposta
    4. **Busca**: ✅ Elasticsearch + filtros MySQL
    5. **Transações**: ✅ Confirmação ACID-compliant
    6. **Rate Limiting**: ✅ Proteção anti-spam
    7. **Observabilidade**: ✅ Logs estruturados + health
    8. **Indexação Automática**: ✅ MySQL → Elasticsearch sync
    9. **Fluxo Completo**: ✅ Flight → Hold → Ticket funcionando 100%

## 🚀 Status do Projeto

**✅ SISTEMA 100% FUNCIONAL E TESTADO**

- Todas as funcionalidades implementadas e validadas
- Fluxo completo de reserva funcionando perfeitamente
- Sincronização automática entre MySQL e Elasticsearch
- Sistema de hold com confirmação de ticket operacional
- PNR gerado automaticamente para tickets confirmados
- Elasticsearch indexando flights, holds e tickets

---

## 📧 Suporte

Para dúvidas ou problemas:

    1. **Instalação rápida**: `make install`
    2. Verifique os logs: `docker-compose logs app`
    3. Execute os testes: `make test`
    4. Verifique URLs: `make urls`


