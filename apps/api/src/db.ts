import { drizzle } from "drizzle-orm/node-postgres";
import pg from "pg";
import * as schema from "@memory-card/db/schema";

const connectionString =
  process.env.DATABASE_URL ??
  "postgres://memory_card:memory_card@localhost:5432/memory_card";

export const pool = new pg.Pool({ connectionString });

export const db = drizzle(pool, { schema });
