# Postman Collection - Airline Booking API

Esta collection contém todos os endpoints da API de Reserva de Voos, incluindo o novo endpoint de criação de voos.

## 📁 Arquivos

- `Airline_Booking_API.postman_collection.json` - Collection principal com todos os endpoints
- `Airline_Booking_API.postman_environment.json` - Environment com variáveis locais

## 🚀 Como Usar

### 1. Importar no Postman

1. Abra o Postman
2. Clique em "Import"
3. Arraste os dois arquivos JSON ou selecione-os
4. Selecione o environment "Airline Booking API - Local"

### 2. Variáveis do Environment

- `baseUrl`: http://localhost:8080/api/v1
- `flight_id`: ID do voo (automaticamente atualizado após criar um voo)
- `user_id`: ID do usuário para operações de hold e tickets
- URLs auxiliares:
  - `mysql_url`: http://localhost:8081 (phpMyAdmin)
  - `elasticsearch_url`: http://localhost:9200
  - `kibana_url`: http://localhost:5601

## 📋 Endpoints Incluídos

### ✅ Health Check
- **GET** `/health` - Verificar status da API

### ✈️ Flights
- **POST** `/flights` - **Criar novo voo** (com indexação automática no Elasticsearch)
- **GET** `/flights/search` - Buscar voos
- **GET** `/flights/{flight_id}/seats` - Ver assentos disponíveis

### 🎫 Holds
- **POST** `/holds` - Criar hold em assento (15 minutos)
- **DELETE** `/holds/{flight_id}/{seat_no}` - Liberar hold

### 🎟️ Tickets
- **POST** `/tickets/confirm` - Confirmar compra de ticket

## 🔄 Fluxo de Teste Recomendado

1. **Health Check** - Verificar se a API está rodando
2. **Create Flight** - Criar um novo voo
3. **Search Flights** - Buscar voos (verificar se o novo voo aparece)
4. **Get Flight Seats** - Ver assentos do voo criado
5. **Create Hold** - Fazer hold em um assento
6. **Confirm Ticket** - Confirmar a compra

## 🧪 Testes Automáticos

Cada request inclui testes automáticos que verificam:
- Status codes corretos
- Estrutura das respostas
- Propriedades obrigatórias
- Atualização automática de variáveis (ex: flight_id)

## 📊 Monitoramento

Após executar os requests, você pode verificar:

- **MySQL**: http://localhost:8081 (phpMyAdmin)
- **Elasticsearch**: http://localhost:9200/flights/_search
- **Kibana**: http://localhost:5601

## 🔧 Exemplo de Criação de Voo

```json
{
  "origin": "SAO",
  "destination": "LAX", 
  "departure_time": "2024-12-15T10:00:00Z",
  "arrival_time": "2024-12-15T20:00:00Z",
  "airline": "TAM",
  "aircraft": "Boeing 737",
  "fare_class": "economy",
  "base_price": 1500.00,
  "seat_config": {
    "economy_rows": 20,
    "business_rows": 5,
    "first_class_rows": 2,
    "seats_per_row": 6
  }
}
```

## ⚡ Features da Collection

- ✅ Variáveis automáticas (flight_id é atualizado após criar voo)
- ✅ Headers automáticos (User-ID, Idempotency-Key)
- ✅ Testes automáticos para validação
- ✅ Exemplos realísticos de dados
- ✅ Environment configurado para desenvolvimento local

## 🛠️ Troubleshooting

Se algum request falhar:

1. Verifique se os serviços estão rodando: `make status`
2. Verifique os logs: `make logs`
3. Reinicie se necessário: `make down && make up`
