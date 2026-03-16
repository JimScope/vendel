import { useQuery } from "@tanstack/react-query"

const ABIS = [
  "arm64-v8a",
  "armeabi-v7a",
  "x86_64",
  "universal",
] as const

export type AndroidAbi = (typeof ABIS)[number]

interface GitHubAsset {
  name: string
  browser_download_url: string
}

interface GitHubRelease {
  tag_name: string
  assets: GitHubAsset[]
}

export function useLatestApks() {
  return useQuery({
    queryKey: ["github", "android-release"],
    queryFn: async (): Promise<Map<string, string>> => {
      const res = await fetch(
        "https://api.github.com/repos/JimScope/vendel-android/releases?per_page=1",
      )
      if (!res.ok) throw new Error("Failed to fetch release")
      const releases: GitHubRelease[] = await res.json()
      if (releases.length === 0) throw new Error("No releases found")
      const release = releases[0]
      const map = new Map<string, string>()
      for (const asset of release.assets) {
        if (!asset.name.endsWith(".apk")) continue
        for (const abi of ABIS) {
          if (asset.name.includes(abi)) {
            map.set(abi, asset.browser_download_url)
          }
        }
      }
      return map
    },
    staleTime: 1000 * 60 * 30,
  })
}
