import { afterAll, beforeAll } from "vitest";

beforeAll(() => {
  if (process.env.DATABASE_URL_TEST) {
    process.env.DATABASE_URL = process.env.DATABASE_URL_TEST;
  }
});

afterAll(() => {
  // Setup global de banco de teste e cleanup.
});
