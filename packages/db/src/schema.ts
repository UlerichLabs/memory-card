import { dificuldades } from "@memory-card/types";
import { sql } from "drizzle-orm";
import {
  boolean,
  check,
  integer,
  json,
  pgEnum,
  pgTable,
  serial,
  text,
  timestamp,
  varchar
} from "drizzle-orm/pg-core";

export const dificuldadeEnum = pgEnum("dificuldade", dificuldades);

export const tipoDesafioEnum = pgEnum("tipo_desafio", [
  "QUANTIDADE",
  "LISTA_COMPLETA",
  "FRANQUIA",
  "EMPRESA",
  "GENERO",
  "MANUAL"
]);

export const tipoFavoritoEnum = pgEnum("tipo_favorito", ["JOGO", "GENERO"]);

export const usuarios = pgTable("usuarios", {
  id: serial("id").primaryKey(),
  nome: varchar("nome", { length: 100 }).notNull(),
  username: varchar("username", { length: 50 }).notNull().unique(),
  email: varchar("email", { length: 255 }).notNull().unique(),
  senhaHash: varchar("senha_hash", { length: 255 }).notNull(),
  avatar: varchar("avatar", { length: 500 }),
  bio: text("bio"),
  redesSociais: json("redes_sociais")
    .$type<{
      youtube?: string;
      instagram?: string;
      twitch?: string;
      x?: string;
    }>()
    .default({}),
  createdAt: timestamp("created_at", { withTimezone: false }).notNull().defaultNow(),
  updatedAt: timestamp("updated_at", { withTimezone: false }).notNull().defaultNow(),
  deletedAt: timestamp("deleted_at", { withTimezone: false })
});

export const jogosZerados = pgTable(
  "jogos_zerados",
  {
    id: serial("id").primaryKey(),
    usuarioId: integer("usuario_id").references(() => usuarios.id).notNull(),
    igdbId: integer("igdb_id"),
    nome: varchar("nome", { length: 300 }).notNull(),
    console: varchar("console", { length: 100 }).notNull(),
    genero: varchar("genero", { length: 100 }),
    tipo: varchar("tipo", { length: 100 }),
    iniciadoEm: timestamp("iniciado_em", { withTimezone: false }),
    finalizadoEm: timestamp("finalizado_em", { withTimezone: false }).notNull(),
    tempoJogado: integer("tempo_jogado").notNull(),
    nota: integer("nota").notNull(),
    dificuldade: dificuldadeEnum("dificuldade").notNull(),
    review: text("review"),
    destaque: boolean("destaque").notNull().default(false),
    igdbCapaUrl: varchar("igdb_capa_url", { length: 500 }),
    igdbDescricao: text("igdb_descricao"),
    createdAt: timestamp("created_at", { withTimezone: false }).notNull().defaultNow(),
    updatedAt: timestamp("updated_at", { withTimezone: false }).notNull().defaultNow(),
    deletedAt: timestamp("deleted_at", { withTimezone: false })
  },
  () => [check("jogos_zerados_nota_range", sql`nota between 1 and 11`)]
);

export const campanhas = pgTable("campanhas", {
  id: serial("id").primaryKey(),
  usuarioId: integer("usuario_id").notNull().references(() => usuarios.id),
  nome: varchar("nome", { length: 200 }).notNull(),
  descricao: text("descricao"),
  createdAt: timestamp("created_at", { withTimezone: false }).defaultNow().notNull(),
  icone: text("icone").default(sql`null`),
  concluida: boolean("concluida").default(false).notNull(),
  deletedAt: timestamp("deleted_at", { withTimezone: false })
});

export const desafios = pgTable("desafios", {
  id: serial("id").primaryKey(),
  usuarioId: integer("usuario_id").notNull().references(() => usuarios.id),
  campanhaId: integer("campanha_id").references(() => campanhas.id),
  nome: varchar("nome", { length: 200 }).notNull(),
  tipo: tipoDesafioEnum("tipo").notNull(),
  filtroValor: varchar("filtro_valor", { length: 200 }),
  meta: integer("meta"),
  dataConclusao: timestamp("data_conclusao", { withTimezone: false }),
  createdAt: timestamp("created_at", { withTimezone: false }).defaultNow().notNull(),
  deletedAt: timestamp("deleted_at", { withTimezone: false })
});

export const desafioItens = pgTable("desafio_itens", {
  id: serial("id").primaryKey(),
  desafioId: integer("desafio_id").notNull().references(() => desafios.id),
  igdbId: integer("igdb_id").notNull(),
  nome: varchar("nome", { length: 300 }).notNull(),
  concluido: boolean("concluido").default(false).notNull(),
  zeramentoId: integer("zeramento_id").references(() => jogosZerados.id)
});

export const conquistas = pgTable("conquistas", {
  id: serial("id").primaryKey(),
  usuarioId: integer("usuario_id").notNull().references(() => usuarios.id),
  desafioId: integer("desafio_id").references(() => desafios.id),
  nome: varchar("nome", { length: 200 }).notNull(),
  descricao: text("descricao"),
  dataConquista: timestamp("data_conquista", { withTimezone: false }).notNull(),
  createdAt: timestamp("created_at", { withTimezone: false }).defaultNow().notNull()
});

export const favoritos = pgTable("favoritos", {
  id: serial("id").primaryKey(),
  usuarioId: integer("usuario_id").notNull().references(() => usuarios.id),
  tipo: tipoFavoritoEnum("tipo").notNull(),
  jogoId: varchar("jogo_id", { length: 200 }).notNull(),
  posicao: integer("posicao").notNull(),
  createdAt: timestamp("created_at", { withTimezone: false }).defaultNow().notNull()
});
