import requests from './requests';

export interface DashboardData {
  web: Record<string, number>;
  ssh: Record<string, number>;
  redis: Record<string, number>;
  mysql: Record<string, number>;
  ftp: Record<string, number>;
  telnet: Record<string, number>;
  memCache: Record<string, number>;
  httpMap: Record<string, number>;
  tftpMap: Record<string, number>;
  vncMap: Record<string, number>;
  esMap: Record<string, number>;
  kubeletMap: Record<string, number>;
  dockerMap: Record<string, number>;
  etcdMap: Record<string, number>;
  apiserverMap: Record<string, number>;
  bashMap: Record<string, number>;
}

export interface PieData {
  regionList: Array<{ name: string; value: string }>;
  ipList: Array<{ name: string; value: string }>;
}

export interface AttackEvent {
  id: number;
  type: string;
  project_name: string;
  agent: string;
  ip: string;
  country: string;
  region: string;
  city: string;
  create_time: string;
  info: string;
  nodetype: string;
  hostname: string;
  agentip: string;
  [key: string]: any;
}

export interface AttackEventData {
  items: AttackEvent[];
  count: string;
}

export interface ApiResponse<T> {
  code: number;
  msg: string;
  data: T;
}

export interface Colony {
  id: number;
  agent_name: string;
  agent_ip: string;
  last_update_time: string;
  es_status: number;
  ftp_status: number;
  http_status: number;
  mem_cache_status: number;
  mysql_status: number;
  redis_status: number;
  ssh_status: number;
  telnet_status: number;
  tftp_status: number;
  vnc_status: number;
  web_status: number;
}

export interface SecretLabel {
  id: number;
  name: string;
  label_type: string;
  file_path: string;
  file_content: string;
  agent_type: string;
  agent_list: string;
  create_time: string;
  update_time: string;
  task_status?: string;
}

export interface SecretLabelAlert {
  id: number;
  secret_label_id: number;
  secret_label_name: string;
  agent: string;
  ip: string;
  access_time: string;
  access_content: string;
  create_time: string;
}

export interface AllAlertsData {
  attack_alerts: AttackEvent[];
  secret_alerts: SecretLabelAlert[];
}

class DashboardApi {
  getDashboardData = async (): Promise<ApiResponse<DashboardData>> => {
    return await requests.get<ApiResponse<DashboardData>>('/get/dashboard/data');
  };

  getPieData = async (): Promise<ApiResponse<PieData>> => {
    return await requests.get<ApiResponse<PieData>>('/get/dashboard/pie_data');
  };

  getAttackEvents = async (pageIndex: number = 1, pageSize: number = 10, type: string = 'all', colony: string = 'all'): Promise<AttackEventData> => {
    return await requests.get<AttackEventData>(`/api/event/paging?pageIndex=${pageIndex}&pageSize=${pageSize}&type=${type}&colony=${colony}`);
  };

  getColonyList = async (): Promise<ApiResponse<Colony[]>> => {
    return await requests.get<ApiResponse<Colony[]>>('/get/colony/list');
  };

  deleteColony = async (id: string): Promise<ApiResponse<any>> => {
    return await requests.postForm<ApiResponse<any>>('/post/colony/del', { id });
  };



  getSecretLabelList = async (): Promise<ApiResponse<SecretLabel[]>> => {
    return await requests.get<ApiResponse<SecretLabel[]>>('/get/secretlabel/list');
  };

  getSecretLabel = async (id: number): Promise<ApiResponse<SecretLabel>> => {
    return await requests.get<ApiResponse<SecretLabel>>(`/get/secretlabel/info?id=${id}`);
  };

  getSecretLabelAlertList = async (): Promise<ApiResponse<SecretLabelAlert[]>> => {
    return await requests.get<ApiResponse<SecretLabelAlert[]>>('/get/secretlabel/alert/list');
  };

  getAllAlerts = async (): Promise<ApiResponse<AllAlertsData>> => {
    return await requests.get<ApiResponse<AllAlertsData>>('/get/all/alerts');
  };

  addSecretLabel = async (data: { name: string; label_type: string; file_path: string; file_content: string; agent_type: string; agent_list: string; monitor_tampering: boolean }): Promise<ApiResponse<any>> => {
    return await requests.postForm<ApiResponse<any>>('/post/secretlabel/add', data);
  };

  updateSecretLabel = async (data: { id: number; name: string; label_type: string; file_path: string; file_content: string; agent_type: string; agent_list: string; monitor_tampering: boolean }): Promise<ApiResponse<any>> => {
    return await requests.postForm<ApiResponse<any>>('/post/secretlabel/update', data);
  };

  deleteSecretLabel = async (id: number): Promise<ApiResponse<any>> => {
    return await requests.postForm<ApiResponse<any>>('/post/secretlabel/del', { id: String(id) });
  };

  getAgentList = async (): Promise<ApiResponse<any>> => {
    return await requests.get<ApiResponse<any>>('/api/v1/agent/list');
  };

  getAgentConfig = async (agentName: string): Promise<ApiResponse<any>> => {
    return await requests.get<ApiResponse<any>>(`/api/v1/agent/config?agent_name=${agentName}`);
  };

  updateAgentConfig = async (config: any): Promise<ApiResponse<any>> => {
    return await requests.post<ApiResponse<any>>('/api/v1/agent/update', config);
  };

  uninstallAgent = async (data: { agent_name: string; agent_ip: string }): Promise<ApiResponse<any>> => {
    return await requests.post<ApiResponse<any>>('/api/v1/agent/uninstall', data);
  };

  login = async (username: string, password: string): Promise<ApiResponse<any>> => {
    return await requests.post<ApiResponse<any>>('/api/login', { username, password });
  };

  logout = async (): Promise<ApiResponse<any>> => {
    return await requests.post<ApiResponse<any>>('/api/logout', {});
  };

  checkLogin = async (): Promise<ApiResponse<any>> => {
    return await requests.get<ApiResponse<any>>('/api/check_login');
  };
}

const dashboardApi = new DashboardApi();
export default dashboardApi;
