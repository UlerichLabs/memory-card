import { existsSync, readFileSync } from "node:fs";
import { resolve } from "node:path";

for (const file of [resolve(process.cwd(), ".env"), resolve(process.cwd(), "../../.env")]) {
  if (!existsSync(file)) {
    continue;
  }

  for (const line of readFileSync(file, "utf8").split(/\r?\n/)) {
    const match = line.match(/^\s*([A-Z0-9_]+)\s*=\s*(.*)\s*$/);

    if (match && process.env[match[1]] === undefined) {
      process.env[match[1]] = match[2].replace(/^["']|["']$/g, "");
    }
  }
}
