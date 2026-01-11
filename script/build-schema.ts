#!/usr/bin/env bun
/// <reference types="bun-types" />
import * as z from "zod"
import { RingOpenCodeConfigSchema } from "../src/config/schema"

const SCHEMA_OUTPUT_PATH = "assets/ring-config.schema.json"

async function main() {
  console.log("Generating JSON Schema...")

  const jsonSchema = z.toJSONSchema(RingOpenCodeConfigSchema, {
    io: "input",
    target: "draft-7",
  })

  const finalSchema = {
    $schema: "http://json-schema.org/draft-07/schema#",
    $id: "https://raw.githubusercontent.com/fredcamaral/ring-for-opencode/main/assets/ring-config.schema.json",
    title: "Ring Configuration",
    description: "Configuration schema for Ring",
    ...jsonSchema,
  }

  await Bun.write(SCHEMA_OUTPUT_PATH, JSON.stringify(finalSchema, null, 2))

  console.log(`Done: ${SCHEMA_OUTPUT_PATH}`)
}

main()
