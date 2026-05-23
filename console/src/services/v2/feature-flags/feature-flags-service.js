import { httpService } from '@/services/v2/base/http/httpService'
import { FeatureFlagsAdapter } from './feature-flags-adapter'

const baseURL = () => import.meta.env.VITE_FLAGS_API_URL

const reqConfig = (clientId) => ({
  baseURL: baseURL(),
  headers: { 'X-Client-Id': clientId },
  withCredentials: false
})

export const featureFlagsService = {
  list: async (clientId) => {
    const { data } = await httpService.request({
      method: 'GET',
      url: 'v1/flags',
      config: reqConfig(clientId)
    })
    return FeatureFlagsAdapter.transformList(data)
  },

  load: async (clientId, key) => {
    const { data } = await httpService.request({
      method: 'GET',
      url: `v1/flags/${key}`,
      config: reqConfig(clientId)
    })
    return FeatureFlagsAdapter.transformItem(key, data)
  },

  create: async (clientId, payload) => {
    const { data } = await httpService.request({
      method: 'POST',
      url: 'v1/flags',
      body: payload,
      config: reqConfig(clientId)
    })
    return FeatureFlagsAdapter.transformItem(payload.key, data)
  },

  patch: async (clientId, key, payload) => {
    const { data } = await httpService.request({
      method: 'PATCH',
      url: `v1/flags/${key}`,
      body: payload,
      config: reqConfig(clientId)
    })
    return FeatureFlagsAdapter.transformItem(key, data)
  },

  remove: async (clientId, key) => {
    await httpService.request({
      method: 'DELETE',
      url: `v1/flags/${key}`,
      config: reqConfig(clientId)
    })
    return 'Flag successfully deleted'
  }
}
