import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { useNavigate } from "@tanstack/react-router"

import pb from "@/lib/pocketbase"
import useCustomToast from "./useCustomToast"

const isLoggedIn = () => {
  return pb.authStore.isValid
}

interface LoginData {
  username: string
  password: string
}

interface SignUpData {
  email: string
  password: string
  passwordConfirm: string
  full_name?: string
}

const useAuth = () => {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const { showErrorToast } = useCustomToast()

  const { data: user } = useQuery({
    queryKey: ["currentUser"],
    queryFn: async () => {
      try {
        const result = await pb.collection("users").authRefresh()
        return result.record
      } catch {
        pb.authStore.clear()
        return null
      }
    },
    enabled: isLoggedIn(),
  })

  const signUpMutation = useMutation({
    mutationFn: async (data: SignUpData) => {
      const record = await pb.collection("users").create({
        email: data.email,
        password: data.password,
        passwordConfirm: data.passwordConfirm,
        full_name: data.full_name || "",
      })
      await pb.collection("users").requestVerification(data.email)
      return { ...record, email: data.email }
    },
    onSuccess: (data) => {
      navigate({ to: "/check-email", search: { email: data?.email || "" } })
    },
    onError: (error: Error) => {
      showErrorToast(error.message || "Registration failed")
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ["users"] })
    },
  })

  const login = async (data: LoginData) => {
    await pb.collection("users").authWithPassword(data.username, data.password)
  }

  const loginMutation = useMutation({
    mutationFn: login,
    onSuccess: () => {
      navigate({ to: "/" })
    },
    onError: (error: Error) => {
      showErrorToast(error.message || "Login failed")
    },
  })

  const logout = () => {
    pb.authStore.clear()
    localStorage.removeItem("access_token")
    queryClient.clear()
    navigate({ to: "/login" })
  }

  return {
    signUpMutation,
    loginMutation,
    logout,
    user,
  }
}

export { isLoggedIn }
export default useAuth
