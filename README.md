# Chef Othon — cardápio digital

Next.js + React + TypeScript no frontend. API Go com SQLite (`modernc.org/sqlite`, sem CGO). Catálogo, sacola persistida no navegador, checkout de pronta entrega e solicitações de orçamento para bolos. Fotos ilustrativas geradas com a ferramenta integrada imagegen, em `frontend/public/images/pastries.png`. Prompt: composição fotográfica 3×2 com fatia de chocolate, cookie, brigadeiros, bolo no pote e dois bolos com frutas, luz natural, sem textos ou interface. Catálogo atual com fotos fornecidas pelo chef: Fatia sabor Kinder e Fatia Ninho a R$ 15,00, Fatia Matilda a R$ 15,00 e Morango cravejado a R$ 15,00.

## Executar localmente

Terminal 1:
```powershell
cd backend
go mod download
go run .
```

Terminal 2:
```powershell
cd frontend
npm ci
npm run dev
```

Acesse a porta informada pelo Next (3000 por padrão; nesta sessão, 3001). A API escuta em 127.0.0.1:8080 e o Next encaminha `/api/*`, sem expor configuração do backend no navegador.

## Verificar
```powershell
cd backend
go test ./...
```
```powershell
cd frontend
npm run build
npx playwright test
```
Os testes de navegador pressupõem ambos os serviços em execução e usam a porta 3001 nesta sessão. Configure `TEST_BASE_URL` se necessário. Instale o navegador de teste com `npx playwright install chromium`.

## Configuração
Backend: `DATABASE_PATH` (padrão `chef-othon.db`, criado no diretório de execução), `API_ADDR` (padrão `127.0.0.1:8080`). Frontend: `API_URL` (padrão `http://127.0.0.1:8080`). O banco e os dados pessoais não são versionados. Para alterar preços iniciais, ajuste o seed antes da primeira execução; bancos existentes preservam seus valores.

## API
- `GET /api/health`: estado do banco.
- `GET /api/products`: catálogo; preços em centavos.
- `POST /api/orders`: pedido com `kind` (`ready` ou `custom`), `name`, `phone`, `method` (`pickup` ou `delivery`), `address`, `notes`, `items` (`productId`, `quantity`). Bolos usam `flavor`, `size`, `date` e lista de itens vazia. Retorna protocolo, subtotal e link do WhatsApp com a mensagem formatada para 5517991540161.

O backend valida entradas, calcula valores a partir do catálogo e grava pedido e itens na mesma transação. Nenhum endpoint público lista dados dos clientes.

## Antes da operação real
Confirmar catálogo, imagens, preços, alergênicos, horários, endereço, taxa e área de entrega, prazos e regras de encomenda. Esta versão registra pedidos no SQLite e abre o WhatsApp com o texto do pedido. O cliente conclui o envio no WhatsApp. Não cobra pagamentos e não tem painel administrativo. Administração autenticada, disponibilidade/estoque, prevenção de duplicidade, proteção contra abuso, backups, privacidade e publicação são etapas seguintes registradas no roadmap.


