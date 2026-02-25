import { useQuery } from "@tanstack/react-query"
import pb from "@/lib/pocketbase"

export function useSMSDetails(id: string, enabled: boolean) {
  return useQuery({
    queryKey: ["sms", id],
    queryFn: async () => {
      return await pb.collection("sms_messages").getOne(id)
    },
    enabled,
    staleTime: 60_000,
  })
}
