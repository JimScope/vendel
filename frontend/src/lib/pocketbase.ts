import PocketBase from "pocketbase"

const pb = new PocketBase(import.meta.env.VITE_API_URL || "http://localhost:8090")

pb.autoCancellation(false)

export default pb
