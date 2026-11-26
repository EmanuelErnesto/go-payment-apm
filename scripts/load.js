// Teste de carga dos 3 endpoints, para gerar volume de traces e spans no Elastic APM.
//
//   k6 run scripts/load.js
//   k6 run -e DURATION=5m -e RATE=50 scripts/load.js
//
// Variáveis: BASE_URL, DURATION, RATE (requisições/s por rota rápida) e
// DELAY_VUS (usuários simultâneos na rota lenta).
//
// Atenção: os POSTs gravam registros reais no banco apontado pela API.
import http from 'k6/http';
import { check } from 'k6';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const DURATION = __ENV.DURATION || '1m';
const RATE = Number(__ENV.RATE || 20);
const DELAY_VUS = Number(__ENV.DELAY_VUS || 5);

const CURRENCIES = ['BRL', 'USD', 'EUR'];
const METHODS = ['credit_card', 'debit_card', 'pix', 'boleto'];

// Taxa fixa de chegada: a carga não cai se a API ficar mais lenta.
function constantRate(exec) {
  return {
    executor: 'constant-arrival-rate',
    exec,
    rate: RATE,
    timeUnit: '1s',
    duration: DURATION,
    preAllocatedVUs: RATE,
    maxVUs: RATE * 5,
  };
}

export const options = {
  scenarios: {
    list: constantRate('listPayments'),
    create: constantRate('createPayment'),
    // Cenário separado para o atraso proposital não segurar as outras rotas.
    delay: {
      executor: 'constant-vus',
      exec: 'createDelayedPayment',
      vus: DELAY_VUS,
      duration: DURATION,
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.01'],
    'http_req_duration{scenario:list}': ['p(95)<500'],
    'http_req_duration{scenario:create}': ['p(95)<500'],
    // Sem limite real: só faz o k6 exibir a latência desta rota separadamente.
    'http_req_duration{scenario:delay}': ['max>=0'],
  },
};

function pick(list) {
  return list[Math.floor(Math.random() * list.length)];
}

function postPayment(path) {
  const body = JSON.stringify({
    amount_cents: 100 + Math.floor(Math.random() * 100000),
    currency: pick(CURRENCIES),
    method: pick(METHODS),
    description: 'teste de carga',
  });

  const res = http.post(`${BASE_URL}${path}`, body, {
    headers: { 'Content-Type': 'application/json' },
  });
  check(res, { 'status 201': (r) => r.status === 201 });
}

export function listPayments() {
  const page = 1 + Math.floor(Math.random() * 5);
  const res = http.get(`${BASE_URL}/payments?page=${page}&page_size=10`, {
    // Agrupa as páginas numa única série de métricas do k6.
    tags: { name: 'GET /payments' },
  });
  check(res, { 'status 200': (r) => r.status === 200 });
}

export function createPayment() {
  postPayment('/payments');
}

export function createDelayedPayment() {
  postPayment('/payments/delay');
}
