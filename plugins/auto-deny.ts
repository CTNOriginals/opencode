import { readFileSync, existsSync } from "fs"

const Toggle = "/tmp/opencode-policy"

export default (async () => {
  return {
    "permission.ask": async (input, output) => {
      if (existsSync(Toggle) && readFileSync(Toggle, "utf-8").trim() === "deny") {
        output.deny = true
      }
    }
  }
})
