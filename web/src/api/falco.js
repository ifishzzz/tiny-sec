import service from '@/utils/request'

const normalizeBaseUrl = (value) => {
  if (!value || value === '/') {
    return ''
  }
  return value.endsWith('/') ? value.slice(0, -1) : value
}

export const getFalcoServerBaseUrl = () => {
  const baseApi = normalizeBaseUrl(import.meta.env.VITE_BASE_API)
  if (typeof window === 'undefined') {
    return baseApi
  }
  return `${window.location.origin}${baseApi}`
}

export const buildFalcoAgentInstallCommand = ({
  enrollKey = '',
  provider = 'aws',
  region = '',
  agentId = ''
} = {}) => {
  const serverUrl = getFalcoServerBaseUrl()
  const scriptBaseUrl = `${serverUrl}/falco/agent/install`
  const safeEnrollKey = enrollKey || '<替换为EnrollKey>'
  const safeRegion = region || '<aws-region>'

  let command = `curl -fsSL "${scriptBaseUrl}/installer" | sudo bash -s -- --server "${serverUrl}" --enroll-key "${safeEnrollKey}" --script-base-url "${scriptBaseUrl}" --provider "${provider}" --region "${safeRegion}"`
  if (agentId) {
    command += ` --agent-id "${agentId}"`
  }
  return command
}

export const getFalcoDashboard = () => {
  return service({
    url: '/falco/dashboard',
    method: 'post'
  })
}

export const getFalcoHostList = (data) => {
  return service({
    url: '/falco/host/list',
    method: 'post',
    data
  })
}

export const getFalcoAgentList = (data) => {
  return service({
    url: '/falco/agent/list',
    method: 'post',
    data
  })
}

export const getFalcoTaskList = (data) => {
  return service({
    url: '/falco/task/list',
    method: 'post',
    data
  })
}

export const createFalcoTask = (data) => {
  return service({
    url: '/falco/task/create',
    method: 'post',
    data
  })
}

export const createFalcoInstallTask = (data) => {
  return service({
    url: '/falco/task/install',
    method: 'post',
    data
  })
}

export const createFalcoUpgradeTask = (data) => {
  return service({
    url: '/falco/task/upgrade',
    method: 'post',
    data
  })
}

export const createFalcoRollbackTask = (data) => {
  return service({
    url: '/falco/task/rollback',
    method: 'post',
    data
  })
}

export const createFalcoReloadTask = (data) => {
  return service({
    url: '/falco/task/reload',
    method: 'post',
    data
  })
}

export const createFalcoRestartTask = (data) => {
  return service({
    url: '/falco/task/restart',
    method: 'post',
    data
  })
}

export const getFalcoRulePackageList = (data) => {
  return service({
    url: '/falco/rule/package/list',
    method: 'post',
    data
  })
}

export const createFalcoRulePackage = (data) => {
  return service({
    url: '/falco/rule/package/create',
    method: 'post',
    data
  })
}

export const publishFalcoRulePackage = (data) => {
  return service({
    url: '/falco/rule/publish',
    method: 'post',
    data
  })
}

export const getFalcoPublishList = (data) => {
  return service({
    url: '/falco/rule/publish/list',
    method: 'post',
    data
  })
}

export const getFalcoEventList = (data) => {
  return service({
    url: '/falco/event/list',
    method: 'post',
    data
  })
}

export const getFalcoSettings = () => {
  return service({
    url: '/falco/settings',
    method: 'get'
  })
}

export const updateFalcoSettings = (data) => {
  return service({
    url: '/falco/settings',
    method: 'put',
    data
  })
}
