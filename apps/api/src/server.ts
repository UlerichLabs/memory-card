import { buildServer } from "./app.js";

const server = await buildServer();

const port = Number(process.env.API_PORT ?? 3333);

try {
  await server.listen({ host: "0.0.0.0", port });
} catch (error) {
  server.log.error(error);
  process.exit(1);
}
