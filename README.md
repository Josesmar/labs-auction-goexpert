# Auction - Backend

Este é o backend do projeto **Auction**, uma plataforma de leilões que permite usuários criarem leilões e fazerem lances. O sistema também permite buscar leilões e determinar o vencedor de cada leilão.

## Índice

- [Pré-requisitos](#pré-requisitos)
- [Instalação](#instalação)
- [Como rodar o projeto em ambiente de desenvolvimento](#como-rodar-o-projeto-em-ambiente-de-desenvolvimento)
- [Estrutura do Projeto](#estrutura-do-projeto)
- [Endpoints](#endpoints)
- [Tecnologias Utilizadas](#tecnologias-utilizadas)
- [Licença](#licença)

---

## Pré-requisitos

Para rodar este projeto em ambiente de desenvolvimento, você precisará de:

- [Docker](https://www.docker.com/get-started)
- [Docker Compose](https://docs.docker.com/compose/install/)
- [Go](https://golang.org/dl/) (caso deseje rodar sem Docker)

---

## Instalação

### 1. Clone o repositório
```bash
git clone https://github.com/seu-usuario/fullcycle-auction-go.git
cd fullcycle-auction-go
```

### 2. Copie o arquivo .env.example para .env

```bash
cp .env.example .env
```

Como rodar o projeto em ambiente de desenvolvimento
1. Configuração do ambiente
Edite o arquivo .env para definir as variáveis necessárias, como MongoDB URL e outras configurações.

2. Inicie os containers com Docker Compose
```bash
docker-compose up --build
```
Isso fará o Docker:

* Criar e subir os containers do MongoDB e da aplicação backend.

A API estará disponível em http://localhost:8080.

### 3.  Rodando diretamente com Go
Caso prefira rodar sem Docker:
1. Instale as dependências do Go:
```bash
go mod tidy
```
2. Compile e execute:
```bash
go run cmd/auction/main.go
```
A API estará disponível em http://localhost:8080.

### Estrutura do Projeto

```bash
.
├── cmd/auction                # Entrypoint da aplicação
│   └── main.go                # Arquivo principal para rodar a aplicação
├── internal/                  # Lógica interna do sistema
│   ├── usecase/               # Casos de uso
│   ├── entity/                # Entidades do sistema
│   ├── repository/            # Repositórios do banco de dados
│   └── infra/                 # Configurações de infraestrutura
├── configuration/             # Configuração do sistema
├── scripts/                   # Scripts auxiliares
├── Dockerfile                 # Dockerfile da aplicação
├── docker-compose.yml         # Arquivo Docker Compose
├── .env                       # Arquivo de variáveis de ambiente
├── README.md                  # Este arquivo
└── go.mod, go.sum             # Dependências do Go
```

### Endpoints
1. Criar um leilão (Auction)
POST /auction

Body:
```bash
{
  "product_name": "Smartphone X",
  "category": "Electronics",
  "description": "Smartphone de última geração",
  "condition": 1
}
```

2. Buscar leilões (Auctions)

```bash
GET /auction
```
Query Parameters:
* status: active ou closed (opcional)
* category: Nome da categoria (opcional)
* productName: Nome do produto (opcional)

3. Buscar leilão por ID

```bash
GET /auction/:auctionId
```
URL Parameters:

* auctionId: ID do leilão

4. Buscar o vencedor do leilão
```bash
GET /auction/winner/:auctionId
```
URL Parameters:

* auctionId: ID do leilão

5. Criar um lance (Bid)
```bash
POST /bid
```

Body:
```bash
{
  "user_id": "123e4567-e89b-12d3-a456-426614174003",
  "auction_id": "123e4567-e89b-12d3-a456-426614174003",
  "amount": 1500
}
```

6. Buscar lances por ID do leilão
```bash
GET /bid/:auctionId
```

URL Parameters:
* auctionId: ID do leilão

7. Criar um usuário (User)

```bash
POST /user
```

Body:
```bash
{
  "id": "123e4567-e89b-12d3-a456-426614174003",
  "name": "John Doe"
}
```

### Tecnologias Utilizadas

* Go (Golang): Backend principal
* MongoDB: Banco de dados NoSQL
* Docker: Containerização do projeto
* Gin: Framework HTTP para API
* Go Modules: Gerenciamento de dependências
* Docker Compose: Orquestração dos containers
