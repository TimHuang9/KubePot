import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import ReactECharts from 'echarts-for-react';
import { Modal, message } from 'antd';
import dashboardApi from '@/api/dashboard-api';

const Dashboard = () => {
  const [activeMenu, setActiveMenu] = useState('overview');
  
  const formatDateTime = (dateTime: string) => {
    if (!dateTime) return '-';
    return dateTime.split('+')[0].split('T').join(' ');
  };
  
  const { data: dashboardData, isLoading: isLoadingData, error: dataError } = useQuery({
    queryKey: ['dashboard/data'],
    queryFn: dashboardApi.getDashboardData,
    refetchInterval: 30000,
    retry: 1,
  });

  const { data: pieData, isLoading: isLoadingPiePie, error: pieError } = useQuery({
    queryKey: ['dashboard/pie_data'],
    queryFn: dashboardApi.getPieData,
    refetchInterval: 30000,
    retry: 1,
  });

  const { data: attackEventData } = useQuery({
    queryKey: ['attack/events/all'],
    queryFn: () => dashboardApi.getAttackEvents(1, 10000, 'all', 'all'),
    refetchInterval: 30000,
    retry: 1,
  });

  const { data: allColonyData } = useQuery({
    queryKey: ['colony/list/all'],
    queryFn: dashboardApi.getColonyList,
    refetchInterval: 30000,
    retry: 1,
  });

  const getLineChartOption = () => {
    if (!dashboardData?.data) {
      return {
        title: {
          text: dataError ? '无法连接到后端服务器' : '暂无数据',
          left: 'center',
          top: 'center',
          textStyle: {
            color: '#999',
            fontSize: 16,
          },
        },
      };
    }

    const hours = Array.from({ length: 24 }, (_, i) => `${i.toString().padStart(2, '0')}:00`);
    
    const data = { ...dashboardData.data };
    
    return {
      tooltip: {
        trigger: 'axis',
      },
      legend: {
        data: ['SSH', 'Redis', 'MySQL', 'FTP', 'Telnet', 'HTTP', 'VNC', 'Kubelet', 'Docker', 'Etcd', 'Bash'],
        top: 0,
      },
      grid: {
        left: '3%',
        right: '4%',
        bottom: '3%',
        containLabel: true,
      },
      xAxis: {
        type: 'category',
        boundaryGap: false,
        data: hours,
      },
      yAxis: {
        type: 'value',
      },
      series: [
        {
          name: 'SSH',
          type: 'line',
          data: hours.map(h => data.ssh?.[h] || 0),
          smooth: true,
        },
        {
          name: 'Redis',
          type: 'line',
          data: hours.map(h => data.redis?.[h] || 0),
          smooth: true,
        },
        {
          name: 'MySQL',
          type: 'line',
          data: hours.map(h => data.mysql?.[h] || 0),
          smooth: true,
        },
        {
          name: 'FTP',
          type: 'line',
          data: hours.map(h => data.ftp?.[h] || 0),
          smooth: true,
        },
        {
          name: 'Telnet',
          type: 'line',
          data: hours.map(h => data.telnet?.[h] || 0),
          smooth: true,
        },
        {
          name: 'HTTP',
          type: 'line',
          data: hours.map(h => data.httpMap?.[h] || 0),
          smooth: true,
        },
        {
          name: 'VNC',
          type: 'line',
          data: hours.map(h => data.vncMap?.[h] || 0),
          smooth: true,
        },
        {
          name: 'Kubelet',
          type: 'line',
          data: hours.map(h => data.kubeletMap?.[h] || 0),
          smooth: true,
        },
        {
          name: 'Docker',
          type: 'line',
          data: hours.map(h => data.dockerMap?.[h] || 0),
          smooth: true,
        },
        {
          name: 'Etcd',
          type: 'line',
          data: hours.map(h => data.etcdMap?.[h] || 0),
          smooth: true,
        },
        {
          name: 'Bash',
          type: 'line',
          data: hours.map(h => data.bashMap?.[h] || 0),
          smooth: true,
        },
      ],
    };
  };

  const getRegionPieOption = () => {
    if (!pieData?.data?.regionList) {
      return {
        title: {
          text: pieError ? '无法连接到后端服务器' : '暂无数据',
          left: 'center',
          top: 'center',
          textStyle: {
            color: '#999',
            fontSize: 16,
          },
        },
      };
    }

    return {
      tooltip: {
        trigger: 'item',
        formatter: '{a} <br/>{b}: {c} ({d}%)',
      },
      legend: {
        orient: 'vertical',
        left: 'left',
      },
      series: [
        {
          name: '攻击地区',
          type: 'pie',
          radius: '50%',
          data: pieData.data.regionList.map(item => ({
            name: item.name,
            value: parseInt(item.value),
          })),
          emphasis: {
            itemStyle: {
              shadowBlur: 10,
              shadowOffsetX: 0,
              shadowColor: 'rgba(0, 0, 0, 0.5)',
            },
          },
        },
      ],
    };
  };

  const getIpPieOption = () => {
    if (!pieData?.data?.ipList) {
      return {
        title: {
          text: pieError ? '无法连接到后端服务器' : '暂无数据',
          left: 'center',
          top: 'center',
          textStyle: {
            color: '#999',
            fontSize: 16,
          },
        },
      };
    }

    return {
      tooltip: {
        trigger: 'item',
        formatter: '{a} <br/>{b}: {c} ({d}%)',
      },
      legend: {
        orient: 'vertical',
        left: 'left',
      },
      series: [
        {
          name: '攻击IP',
          type: 'pie',
          radius: '50%',
          data: pieData.data.ipList.map(item => ({
            name: item.name,
            value: parseInt(item.value),
          })),
          emphasis: {
            itemStyle: {
              shadowBlur: 10,
              shadowOffsetX: 0,
              shadowColor: 'rgba(0, 0, 0, 0.5)',
            },
          },
        },
      ],
    };
  };

  const getTotalAttacks = () => {
    return parseInt(attackEventData?.count || '0');
  };

  const getActiveServices = () => {
    if (!dashboardData?.data) return 0;
    
    const data = { ...dashboardData.data };
    let active = 0;
    
    const services = ['ssh', 'redis', 'mysql', 'ftp', 'telnet', 'httpMap', 'vncMap', 'esMap', 'kubeletMap', 'dockerMap', 'etcdMap', 'apiserverMap', 'bashMap'];
    
    services.forEach(service => {
      const serviceData = data[service as keyof typeof data];
      if (serviceData && typeof serviceData === 'object') {
        const hasActivity = Object.values(serviceData).some(val => Number(val) > 0);
        if (hasActivity) active++;
      }
    });
    
    return active;
  };

  if (isLoadingData || isLoadingPiePie) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-2xl text-gray-600">加载中...</div>
      </div>
    );
  }

  const MonitorPanel = () => {
    const colonies = (allColonyData?.data || []).filter(colony => colony.agent_name && colony.agent_name.trim() !== '');
    
    // 计算时间差，返回多少分钟前
    const getMinutesAgo = (timeString: string) => {
      const lastUpdate = new Date(timeString);
      const now = new Date();
      const diffMs = now.getTime() - lastUpdate.getTime();
      const diffMins = Math.round(diffMs / 60000);
      return diffMins;
    };

    // 判断节点是否在线（基于最后更新时间，超过5分钟视为离线）
    const isNodeOnline = (lastUpdateTime: string) => {
      const minsAgo = getMinutesAgo(lastUpdateTime);
      return minsAgo <= 5;
    };
    
    const onlineCount = colonies.filter(colony => 
      isNodeOnline(colony.last_update_time)
    ).length;
    const offlineCount = colonies.length - onlineCount;
    const onlinePercent = colonies.length > 0 ? Math.round((onlineCount / colonies.length) * 100) : 0;

    return (
    <div className="container mx-auto px-6 py-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-800 mb-2">监控面板</h1>
        <p className="text-gray-600">实时监控蜜罐攻击情况</p>
      </div>
      
      {dataError && (
        <div className="mb-4 p-4 bg-yellow-100 border border-yellow-400 text-yellow-700 rounded">
          ⚠️ 无法连接到后端服务器，请确保KubePot后端服务正在运行（端口9001）
        </div>
      )}

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        <div className="stat-card">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-500 text-sm mb-1">总攻击次数</p>
              <p className="text-3xl font-bold text-blue-600">{getTotalAttacks()}</p>
            </div>
            <div className="w-12 h-12 bg-blue-100 rounded-full flex items-center justify-center">
              <svg className="w-6 h-6 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
          </div>
        </div>

        <div className="stat-card">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-500 text-sm mb-1">在线终端数</p>
              <p className="text-3xl font-bold text-green-600">{onlineCount}</p>
            </div>
            <div className="w-12 h-12 bg-green-100 rounded-full flex items-center justify-center">
              <svg className="w-6 h-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01" />
              </svg>
            </div>
          </div>
        </div>

        <div className="stat-card">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-500 text-sm mb-1">离线终端数</p>
              <p className="text-3xl font-bold text-red-600">{offlineCount}</p>
            </div>
            <div className="w-12 h-12 bg-red-100 rounded-full flex items-center justify-center">
              <svg className="w-6 h-6 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </div>
          </div>
        </div>

        <div className="stat-card">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-500 text-sm mb-1">在线率</p>
              <p className="text-3xl font-bold text-purple-600">{onlinePercent}%</p>
            </div>
            <div className="w-12 h-12 bg-purple-100 rounded-full flex items-center justify-center">
              <svg className="w-6 h-6 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
              </svg>
            </div>
          </div>
        </div>
      </div>

      <div className="card mb-8">
        <h2 className="text-xl font-semibold text-gray-800 mb-4">24小时攻击趋势</h2>
        <ReactECharts option={getLineChartOption()} style={{ height: '400px' }} />
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
        <div className="card">
          <h2 className="text-xl font-semibold text-gray-800 mb-4">攻击地区分布</h2>
          <ReactECharts option={getRegionPieOption()} style={{ height: '400px' }} />
        </div>

        <div className="card">
          <h2 className="text-xl font-semibold text-gray-800 mb-4">攻击IP排行</h2>
          <ReactECharts option={getIpPieOption()} style={{ height: '400px' }} />
        </div>
      </div>
    </div>
  );
  }
  const MainContent = () => {
    const [selectedEvent, setSelectedEvent] = useState<any>(null);
    const [currentPage, setCurrentPage] = useState(1);
    const [jumpPage, setJumpPage] = useState('');
    const queryClient = useQueryClient();
    const pageSize = 10;

    const { data: attackEventData, isLoading: isLoadingAttackEvents, error: attackEventsError, refetch: refetchAttackEvents } = useQuery({
      queryKey: ['attack/events', currentPage],
      queryFn: () => dashboardApi.getAttackEvents(currentPage, pageSize, 'all', 'all'),
      refetchInterval: 30000,
      retry: 1,
    });

    const { data: colonyData, isLoading: isLoadingColony, error: colonyError, refetch: refetchColony } = useQuery({
      queryKey: ['colony/list'],
      queryFn: dashboardApi.getColonyList,
      refetchInterval: 30000,
      retry: 1,
    });

    const { data: allAlertsData, isLoading: isLoadingAllAlerts, error: allAlertsError, refetch: refetchAllAlerts } = useQuery({
      queryKey: ['all/alerts'],
      queryFn: dashboardApi.getAllAlerts,
      refetchInterval: 30000,
      retry: 1,
    });

    const { data: secretLabelAlertData, isLoading: isLoadingSecretLabelAlerts, error: secretLabelAlertsError, refetch: refetchSecretLabelAlerts } = useQuery({
      queryKey: ['secretlabel/alert/list'],
      queryFn: dashboardApi.getSecretLabelAlertList,
      refetchInterval: 30000,
      retry: 1,
    });

    const deleteColonyMutation = useMutation({
      mutationFn: (id: number) => dashboardApi.deleteColony(String(id)),
      onSuccess: () => {
        queryClient.invalidateQueries({ queryKey: ['colony/list'] });
      },
    });

    const handleDeleteColony = (id: number) => {
      Modal.confirm({
        title: '删除节点',
        content: '确定要删除该节点吗？',
        okText: '确定',
        cancelText: '取消',
        onOk: () => {
          deleteColonyMutation.mutate(id);
          message.success('节点已删除');
        }
      });
    };

    const handleUninstallAgent = async (agentName: string, agentIp: string) => {
      Modal.confirm({
        title: '卸载节点',
        content: `确定要卸载节点 ${agentName} 吗？`,
        okText: '确定',
        cancelText: '取消',
        onOk: async () => {
          try {
            const response = await dashboardApi.uninstallAgent({ agent_name: agentName, agent_ip: agentIp });
            if (response.code === 0 || response.code === 200) {
              message.success('卸载命令已发送，请等待执行完成');
              refetchColony();
            } else {
              message.error('卸载失败: ' + response.msg);
            }
          } catch (error) {
            console.error('卸载节点失败:', error);
            message.error('卸载失败，请检查网络连接');
          }
        }
      });
    };

    const attackEvents = attackEventData?.items || [];
    const totalCount = attackEventData?.count || '0';
    const colonies = (colonyData?.data || []).filter(colony => colony.agent_name && colony.agent_name.trim() !== '');

    const getAttackBehavior = (event: any) => {
      const type = event.type?.toLowerCase() || '';
      const info = event.info || '';
      
      if (type.includes('kubelet')) {
        if (info.includes('pod') || info.includes('pods')) {
          return '获取Pod列表';
        } else if (info.includes('container') || info.includes('containers')) {
          return '获取容器信息';
        } else if (info.includes('node') || info.includes('nodes')) {
          return '获取节点信息';
        } else if (info.includes('logs')) {
          return '查看日志';
        } else {
          return 'Kubelet API访问';
        }
      } else if (type.includes('docker')) {
        if (info.includes('ps') || info.includes('list')) {
          return '列出容器';
        } else if (info.includes('info')) {
          return '获取Docker信息';
        } else if (info.includes('exec')) {
          return '执行容器命令';
        } else {
          return 'Docker API访问';
        }
      } else if (type.includes('etcd')) {
        if (info.includes('get') || info.includes('read')) {
          return '读取配置数据';
        } else if (info.includes('put') || info.includes('write')) {
          return '写入配置数据';
        } else {
          return 'Etcd数据访问';
        }
      } else if (type.includes('apiserver')) {
        if (info.includes('pod') || info.includes('pods')) {
          return '访问Pod资源';
        } else if (info.includes('secret')) {
          return '访问Secret';
        } else if (info.includes('configmap')) {
          return '访问ConfigMap';
        } else {
          return 'Apiserver API访问';
        }
      } else if (type.includes('bash')) {
        if (info.includes('rm') || info.includes('delete')) {
          return '执行删除命令';
        } else if (info.includes('cat') || info.includes('read')) {
          return '读取文件';
        } else if (info.includes('wget') || info.includes('curl')) {
          return '下载文件';
        } else {
          return '执行Shell命令';
        }
      } else if (type.includes('ssh')) {
        return 'SSH登录尝试';
      } else if (type.includes('redis')) {
        return 'Redis访问';
      } else if (type.includes('mysql')) {
        return 'MySQL访问';
      } else if (type.includes('ftp')) {
        return 'FTP访问';
      } else if (type.includes('http')) {
        return 'HTTP访问';
      } else {
        return '未知攻击行为';
      }
    };

    const getRiskLevel = (event: any) => {
      return '高危';
    };

    const getRiskLevelClass = (level: string) => {
      switch (level) {
        case '高危':
          return 'text-red-600';
        case '中危':
          return 'text-orange-600';
        case '低危':
          return 'text-yellow-600';
        default:
          return 'text-gray-600';
      }
    };

    const getRiskStats = () => {
      const stats = {
        high: 0,
        medium: 0,
        low: 0
      };

      attackEvents.forEach(event => {
        const level = getRiskLevel(event);
        if (level === '高危') {
          stats.high++;
        } else if (level === '中危') {
          stats.medium++;
        } else {
          stats.low++;
        }
      });

      return stats;
    };

    const overviewItems = [
      { key: 'overview', label: '监控面板' },
      { key: 'alert-overview', label: '告警总览' }
    ];

    const threatItems = [
      { key: 'attack-list', label: '攻击列表' }
    ];

    const attackerItems = [
      { key: 'attacker-profile', label: '攻击者画像' }
    ];

    const countermeasureItems = [
      { key: 'countermeasure-config', label: '反制配置' },
      { key: 'traceability-config', label: '溯源配置' }
    ];

    const secretItems = [
      { key: 'secret-label-alert', label: '密标告警' },
      { key: 'secret-label', label: '密标管理' }
    ];

    const SecretLabelContent = () => {
      const [showAddModal, setShowAddModal] = useState(false);
      const [showEditModal, setShowEditModal] = useState(false);
      const [editingLabel, setEditingLabel] = useState<any>(null);
      const [formData, setFormData] = useState({
        name: '',
        label_type: 'kubeconfig',
        file_path: '',
        file_content: '',
        agent_type: 'all',
        agent_list: '',
        monitor_tampering: false
      });
      const [selectedLabels, setSelectedLabels] = useState<number[]>([]);
      const [selectAll, setSelectAll] = useState(false);
      const queryClient = useQueryClient();

      const { data: secretLabelData, isLoading: isLoadingSecretLabels, error: secretLabelsError, refetch: refetchSecretLabels } = useQuery({
        queryKey: ['secretlabel/list'],
        queryFn: dashboardApi.getSecretLabelList,
        refetchInterval: 30000,
        retry: 1,
      });

      const { data: colonyData } = useQuery({
        queryKey: ['colony/list'],
        queryFn: dashboardApi.getColonyList,
        refetchInterval: 30000,
        retry: 1,
      });

      const getMinutesAgo = (timeString: string) => {
        const lastUpdate = new Date(timeString);
        const now = new Date();
        const diffMs = now.getTime() - lastUpdate.getTime();
        const diffMins = Math.round(diffMs / 60000);
        return diffMins;
      };

      const isNodeOnline = (lastUpdateTime: string) => {
        const minsAgo = getMinutesAgo(lastUpdateTime);
        return minsAgo <= 5;
      };

      const colonies = (colonyData?.data || []).filter(colony => colony.agent_name && colony.agent_name.trim() !== '');
      const onlineColonies = colonies.filter(colony => isNodeOnline(colony.last_update_time));

      const selectedAgents = formData.agent_list ? formData.agent_list.split(',').filter(Boolean) : [];

      const handleAgentToggle = (agentName: string) => {
        let newSelectedAgents: string[];
        if (selectedAgents.includes(agentName)) {
          newSelectedAgents = selectedAgents.filter(a => a !== agentName);
        } else {
          newSelectedAgents = [...selectedAgents, agentName];
        }
        setFormData({ ...formData, agent_list: newSelectedAgents.join(',') });
      };

      const secretLabels = secretLabelData?.data || [];

      const handleAdd = async () => {
        try {
          await dashboardApi.addSecretLabel(formData);
          setShowAddModal(false);
          setFormData({ name: '', label_type: 'kubeconfig', file_path: '', file_content: '', agent_type: 'all', agent_list: '', monitor_tampering: false });
          queryClient.invalidateQueries({ queryKey: ['secretlabel/list'] });
          message.success('添加成功');
        } catch (error) {
          message.error('添加失败');
        }
      };

      const handleEdit = async () => {
        if (!editingLabel) return;
        try {
          await dashboardApi.updateSecretLabel({ ...formData, id: editingLabel.id });
          setShowEditModal(false);
          setEditingLabel(null);
          setFormData({ name: '', label_type: 'kubeconfig', file_path: '', file_content: '', agent_type: 'all', agent_list: '', monitor_tampering: false });
          queryClient.invalidateQueries({ queryKey: ['secretlabel/list'] });
          message.success('更新成功');
        } catch (error) {
          message.error('更新失败');
        }
      };

      const handleDelete = async (id: number) => {
        Modal.confirm({
          title: '删除密标',
          content: '确定要删除该密标吗？',
          okText: '确定',
          cancelText: '取消',
          onOk: async () => {
            try {
              await dashboardApi.deleteSecretLabel(id);
              queryClient.invalidateQueries({ queryKey: ['secretlabel/list'] });
              message.success('删除成功');
            } catch (error) {
              message.error('删除失败');
            }
          }
        });
      };

      const openEditModal = (label: any) => {
        setEditingLabel(label);
        setFormData({
          name: label.name,
          label_type: label.label_type,
          file_path: label.file_path,
          file_content: label.file_content,
          agent_type: label.agent_type || 'all',
          agent_list: label.agent_list || '',
          monitor_tampering: label.monitor_tampering || false
        });
        setShowEditModal(true);
      };

      const handleSelectLabel = (id: number) => {
        setSelectedLabels(prev => {
          if (prev.includes(id)) {
            return prev.filter(item => item !== id);
          } else {
            return [...prev, id];
          }
        });
      };

      const handleSelectAll = () => {
        if (selectAll) {
          setSelectedLabels([]);
          setSelectAll(false);
        } else {
          const allIds = secretLabels.map(label => label.id);
          setSelectedLabels(allIds);
          setSelectAll(true);
        }
      };

      const handleBatchDelete = async () => {
        if (selectedLabels.length === 0) {
          message.warning('请选择要删除的密标');
          return;
        }

        Modal.confirm({
          title: '批量删除密标',
          content: `确定要删除选中的 ${selectedLabels.length} 个密标吗？`,
          okText: '确定',
          cancelText: '取消',
          onOk: async () => {
            try {
              for (const id of selectedLabels) {
                await dashboardApi.deleteSecretLabel(id);
              }
              setSelectedLabels([]);
              setSelectAll(false);
              queryClient.invalidateQueries({ queryKey: ['secretlabel/list'] });
              message.success('批量删除成功');
            } catch (error) {
              message.error('批量删除失败');
            }
          }
        });
      };

      return (
        <div className="space-y-6">
          {secretLabelsError && (
            <div className="mb-4 p-4 bg-yellow-100 border border-yellow-400 text-yellow-700 rounded">
              ⚠️ 无法连接到后端服务器，请确保KubePot后端服务正在运行（端口9001）
            </div>
          )}

          {isLoadingSecretLabels && (
            <div className="flex items-center justify-center py-8">
              <div className="text-gray-600">加载中...</div>
            </div>
          )}

          <div className="bg-white rounded-lg p-6 shadow-sm">
            <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center mb-4 gap-3">
              <h3 className="text-lg font-semibold text-gray-800">密标列表</h3>
              <div className="flex items-center gap-3 w-full sm:w-auto">
                {secretLabels.length > 0 && (
                  <button
                    onClick={handleBatchDelete}
                    className={`px-4 py-2 rounded hover:shadow transition-all duration-200 ${selectedLabels.length > 0 ? 'bg-red-500 text-white hover:bg-red-600' : 'bg-gray-200 text-gray-500 cursor-not-allowed'}`}
                    disabled={selectedLabels.length === 0}
                  >
                    批量删除
                    {selectedLabels.length > 0 && (
                      <span className="ml-2 bg-white text-red-500 text-xs px-2 py-1 rounded-full">
                        {selectedLabels.length}
                      </span>
                    )}
                  </button>
                )}
                <button
                  onClick={() => setShowAddModal(true)}
                  className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 hover:shadow transition-all duration-200"
                >
                  新增密标
                </button>
              </div>
            </div>
            <div className="overflow-x-auto">
              <table className="w-full border-collapse">
                <thead>
                  <tr className="bg-gray-50">
                    <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700 border-b">
                      <input
                        type="checkbox"
                        checked={selectAll}
                        onChange={handleSelectAll}
                        className="w-4 h-4 text-blue-600 rounded focus:ring-blue-500"
                      />
                    </th>
                    <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700 border-b">密标名称</th>
                    <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700 border-b">密标类型</th>
                    <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700 border-b">文件路径</th>
                    <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700 border-b">Agent选择</th>
                    <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700 border-b">任务状态</th>
                    <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700 border-b">创建时间</th>
                    <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700 border-b">更新时间</th>
                    <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700 border-b">操作</th>
                  </tr>
                </thead>
                <tbody>
                  {secretLabels.map((label, index) => (
                    <tr key={label.id} className={index % 2 === 0 ? 'bg-white' : 'bg-gray-50'}>
                      <td className="px-4 py-3 text-sm border-b">
                        <input
                          type="checkbox"
                          checked={selectedLabels.includes(label.id)}
                          onChange={() => handleSelectLabel(label.id)}
                          className="w-4 h-4 text-blue-600 rounded focus:ring-blue-500"
                        />
                      </td>
                      <td className="px-4 py-3 text-sm text-gray-700 border-b">{label.name}</td>
                      <td className="px-4 py-3 text-sm text-gray-700 border-b">{label.label_type}</td>
                      <td className="px-4 py-3 text-sm text-gray-700 border-b max-w-xs truncate" title={label.file_path}>{label.file_path}</td>
                      <td className="px-4 py-3 text-sm text-gray-700 border-b">
                        {label.agent_type === 'all' ? (
                          <span className="px-2 py-1 bg-blue-100 text-blue-800 rounded-full text-xs font-medium">全部Agent</span>
                        ) : (
                          <div className="flex flex-wrap gap-1">
                            {label.agent_list ? label.agent_list.split(',').filter(Boolean).map((agent, i) => (
                              <span key={i} className="px-2 py-1 bg-green-100 text-green-800 rounded-full text-xs font-medium">
                                {agent}
                              </span>
                            )) : (
                              <span className="text-gray-400 text-xs">未指定</span>
                            )}
                          </div>
                        )}
                      </td>
                      <td className="px-4 py-3 text-sm border-b">
                        {label.task_status === 'pending' ? (
                          <span className="px-2 py-1 bg-yellow-100 text-yellow-800 rounded-full text-xs font-medium">待接收</span>
                        ) : label.task_status === 'completed' ? (
                          <span className="px-2 py-1 bg-green-100 text-green-800 rounded-full text-xs font-medium">已放置</span>
                        ) : label.task_status === 'failed' ? (
                          <span className="px-2 py-1 bg-red-100 text-red-800 rounded-full text-xs font-medium">失败</span>
                        ) : (
                          <span className="px-2 py-1 bg-gray-100 text-gray-800 rounded-full text-xs font-medium">未知</span>
                        )}
                      </td>
                      <td className="px-4 py-3 text-sm text-gray-700 border-b">{formatDateTime(label.create_time)}</td>
                      <td className="px-4 py-3 text-sm text-gray-700 border-b">{formatDateTime(label.update_time)}</td>
                      <td className="px-4 py-3 text-sm border-b space-x-2">
                        <button
                          onClick={() => openEditModal(label)}
                          className="text-blue-600 hover:text-blue-800 font-medium"
                        >
                          编辑
                        </button>
                        <button
                          onClick={() => handleDelete(label.id)}
                          className="text-red-600 hover:text-red-800 font-medium"
                        >
                          删除
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            {secretLabels.length === 0 && (
              <div className="text-center py-8 text-gray-500">
                暂无密标数据
              </div>
            )}
          </div>

          {showAddModal && (
            <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
              <div className="bg-white rounded-lg p-6 max-w-2xl w-full mx-4 shadow-xl">
                <div className="flex justify-between items-center mb-4">
                  <h3 className="text-lg font-semibold text-gray-800">新增密标</h3>
                  <button
                    onClick={() => {
                      setShowAddModal(false);
                      setFormData({ name: '', label_type: 'kubeconfig', file_path: '', file_content: '', agent_type: 'all', agent_list: '', monitor_tampering: false });
                    }}
                    className="text-gray-500 hover:text-gray-700"
                  >
                    <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                    </svg>
                  </button>
                </div>
                <div className="space-y-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">密标名称</label>
                    <input
                      type="text"
                      value={formData.name}
                      onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                      className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                      placeholder="请输入密标名称"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">密标类型</label>
                    <select
                      value={formData.label_type}
                      onChange={(e) => setFormData({ ...formData, label_type: e.target.value })}
                      className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    >
                      <option value="kubeconfig">kubeconfig</option>
                      <option value="custom">自定义</option>
                    </select>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">文件路径</label>
                    <input
                      type="text"
                      value={formData.file_path}
                      onChange={(e) => setFormData({ ...formData, file_path: e.target.value })}
                      className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                      placeholder="请输入文件路径，例如 /root/.kube/config"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">文件内容</label>
                    <textarea
                      value={formData.file_content}
                      onChange={(e) => setFormData({ ...formData, file_content: e.target.value })}
                      className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                      rows={8}
                      placeholder="请输入文件内容"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Agent选择</label>
                    <select
                      value={formData.agent_type}
                      onChange={(e) => setFormData({ ...formData, agent_type: e.target.value })}
                      className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    >
                      <option value="all">全部Agent</option>
                      <option value="specific">指定Agent</option>
                    </select>
                  </div>
                  {formData.agent_type === 'specific' && (
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">选择在线Agent</label>
                      {onlineColonies.length === 0 ? (
                        <div className="p-4 bg-gray-50 border border-gray-200 rounded-md text-gray-500 text-sm">
                          暂无在线Agent
                        </div>
                      ) : (
                        <div className="max-h-48 overflow-y-auto border border-gray-300 rounded-md p-2 space-y-2">
                          {onlineColonies.map((colony) => (
                            <label 
                              key={colony.id} 
                              className="flex items-center gap-2 cursor-pointer hover:bg-blue-50 p-2 rounded-lg transition-all duration-200 group border border-transparent hover:border-blue-200"
                              title={`Agent名称: ${colony.agent_name}\nIP地址: ${colony.agent_ip || '未知'}\n最后更新: ${colony.last_update_time}`}
                            >
                              <input
                                type="checkbox"
                                checked={selectedAgents.includes(colony.agent_name)}
                                onChange={() => handleAgentToggle(colony.agent_name)}
                                className="w-4 h-4 text-blue-600 rounded focus:ring-blue-500"
                              />
                              <div className="flex-1">
                                <div className="text-sm font-medium text-gray-800">{colony.agent_name}</div>
                                <div className="text-xs text-gray-500">{colony.agent_ip || '未知IP'}</div>
                              </div>
                              <span className="px-2 py-1 bg-green-100 text-green-800 text-xs rounded-full font-medium">在线</span>
                            </label>
                          ))}
                        </div>
                      )}
                      <div className="mt-2 text-xs text-gray-500">
                        已选择 {selectedAgents.length} 个Agent
                      </div>
                    </div>
                  )}
                  <div>
                    <label className="relative inline-flex items-center cursor-pointer">
                      <input 
                        type="checkbox" 
                        className="sr-only peer"
                        checked={formData.monitor_tampering}
                        onChange={(e) => setFormData({ ...formData, monitor_tampering: e.target.checked })}
                      />
                      <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-500"></div>
                      <span className="ml-3 text-sm font-medium text-gray-700">监控文件篡改</span>
                    </label>
                  </div>
                  <div className="flex justify-end gap-2">
                    <button
                      onClick={() => {
                        setShowAddModal(false);
                        setFormData({ name: '', label_type: 'kubeconfig', file_path: '', file_content: '', agent_type: 'all', agent_list: '', monitor_tampering: false });
                      }}
                      className="px-4 py-2 bg-gray-200 text-gray-800 rounded hover:bg-gray-300"
                    >
                      取消
                    </button>
                    <button
                      onClick={handleAdd}
                      className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600"
                    >
                      确定
                    </button>
                  </div>
                </div>
              </div>
            </div>
          )}

          {showEditModal && editingLabel && (
            <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
              <div className="bg-white rounded-lg p-6 max-w-2xl w-full mx-4 shadow-xl">
                <div className="flex justify-between items-center mb-4">
                  <h3 className="text-lg font-semibold text-gray-800">编辑密标</h3>
                  <button
                    onClick={() => {
                      setShowEditModal(false);
                      setEditingLabel(null);
                      setFormData({ name: '', label_type: 'kubeconfig', file_path: '', file_content: '', agent_type: 'all', agent_list: '', monitor_tampering: false });
                    }}
                    className="text-gray-500 hover:text-gray-700"
                  >
                    <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                    </svg>
                  </button>
                </div>
                <div className="space-y-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">密标名称</label>
                    <input
                      type="text"
                      value={formData.name}
                      onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                      className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                      placeholder="请输入密标名称"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">密标类型</label>
                    <select
                      value={formData.label_type}
                      onChange={(e) => setFormData({ ...formData, label_type: e.target.value })}
                      className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    >
                      <option value="kubeconfig">kubeconfig</option>
                      <option value="custom">自定义</option>
                    </select>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">文件路径</label>
                    <input
                      type="text"
                      value={formData.file_path}
                      onChange={(e) => setFormData({ ...formData, file_path: e.target.value })}
                      className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                      placeholder="请输入文件路径，例如 /root/.kube/config"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">文件内容</label>
                    <textarea
                      value={formData.file_content}
                      onChange={(e) => setFormData({ ...formData, file_content: e.target.value })}
                      className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                      rows={8}
                      placeholder="请输入文件内容"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">Agent选择</label>
                    <select
                      value={formData.agent_type}
                      onChange={(e) => setFormData({ ...formData, agent_type: e.target.value })}
                      className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    >
                      <option value="all">全部Agent</option>
                      <option value="specific">指定Agent</option>
                    </select>
                  </div>
                  {formData.agent_type === 'specific' && (
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">选择在线Agent</label>
                      {onlineColonies.length === 0 ? (
                        <div className="p-4 bg-gray-50 border border-gray-200 rounded-md text-gray-500 text-sm">
                          暂无在线Agent
                        </div>
                      ) : (
                        <div className="max-h-48 overflow-y-auto border border-gray-300 rounded-md p-2 space-y-2">
                          {onlineColonies.map((colony) => (
                            <label 
                              key={colony.id} 
                              className="flex items-center gap-2 cursor-pointer hover:bg-blue-50 p-2 rounded-lg transition-all duration-200 group border border-transparent hover:border-blue-200"
                              title={`Agent名称: ${colony.agent_name}\nIP地址: ${colony.agent_ip || '未知'}\n最后更新: ${colony.last_update_time}`}
                            >
                              <input
                                type="checkbox"
                                checked={selectedAgents.includes(colony.agent_name)}
                                onChange={() => handleAgentToggle(colony.agent_name)}
                                className="w-4 h-4 text-blue-600 rounded focus:ring-blue-500"
                              />
                              <div className="flex-1">
                                <div className="text-sm font-medium text-gray-800">{colony.agent_name}</div>
                                <div className="text-xs text-gray-500">{colony.agent_ip || '未知IP'}</div>
                              </div>
                              <span className="px-2 py-1 bg-green-100 text-green-800 text-xs rounded-full font-medium">在线</span>
                            </label>
                          ))}
                        </div>
                      )}
                      <div className="mt-2 text-xs text-gray-500">
                        已选择 {selectedAgents.length} 个Agent
                      </div>
                    </div>
                  )}
                  <div>
                    <label className="relative inline-flex items-center cursor-pointer">
                      <input 
                        type="checkbox" 
                        className="sr-only peer"
                        checked={formData.monitor_tampering}
                        onChange={(e) => setFormData({ ...formData, monitor_tampering: e.target.checked })}
                      />
                      <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-500"></div>
                      <span className="ml-3 text-sm font-medium text-gray-700">监控文件篡改</span>
                    </label>
                  </div>
                  <div className="flex justify-end gap-2">
                    <button
                      onClick={() => {
                        setShowEditModal(false);
                        setEditingLabel(null);
                        setFormData({ name: '', label_type: 'kubeconfig', file_path: '', file_content: '', agent_type: 'all', agent_list: '', monitor_tampering: false });
                      }}
                      className="px-4 py-2 bg-gray-200 text-gray-800 rounded hover:bg-gray-300"
                    >
                      取消
                    </button>
                    <button
                      onClick={handleEdit}
                      className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600"
                    >
                      确定
                    </button>
                  </div>
                </div>
              </div>
            </div>
          )}
        </div>
      );
    };

    const SecretLabelAlertContent = () => {
      const secretLabelAlerts = secretLabelAlertData?.data || [];

      return (
        <div className="space-y-6">
          {secretLabelAlertsError && (
            <div className="mb-4 p-4 bg-yellow-100 border border-yellow-400 text-yellow-700 rounded">
              ⚠️ 无法连接到后端服务器，请确保KubePot后端服务正在运行（端口9001）
            </div>
          )}

          {isLoadingSecretLabelAlerts && (
            <div className="flex items-center justify-center py-8">
              <div className="text-gray-600">加载中...</div>
            </div>
          )}

          <div className="bg-white rounded-lg p-6 shadow-sm">
            <div className="flex justify-between items-center mb-4">
              <h3 className="text-lg font-semibold text-gray-800">密标告警列表</h3>
              <div className="flex items-center gap-2">
                <span className="text-sm text-gray-500">自动刷新：30秒</span>
                <button
                  onClick={() => refetchSecretLabelAlerts()}
                  className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600"
                >
                  刷新
                </button>
              </div>
            </div>
            <div className="overflow-x-auto">
              <table className="w-full border-collapse">
                <thead>
                  <tr className="bg-gray-50">
                    <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700 border-b">密标名称</th>
                    <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700 border-b">Agent</th>
                    <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700 border-b">IP地址</th>
                    <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700 border-b">访问时间</th>
                    <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700 border-b">访问内容</th>
                    <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700 border-b">告警时间</th>
                  </tr>
                </thead>
                <tbody>
                  {secretLabelAlerts.map((alert, index) => (
                    <tr key={alert.id} className={index % 2 === 0 ? 'bg-white' : 'bg-gray-50'}>
                      <td className="px-4 py-3 text-sm text-gray-700 border-b">{alert.secret_label_name}</td>
                      <td className="px-4 py-3 text-sm text-gray-700 border-b">{alert.agent}</td>
                      <td className="px-4 py-3 text-sm text-gray-700 border-b">{alert.ip}</td>
                      <td className="px-4 py-3 text-sm text-gray-700 border-b">{formatDateTime(alert.access_time)}</td>
                      <td className="px-4 py-3 text-sm text-gray-700 border-b max-w-xs truncate" title={alert.access_content}>{alert.access_content}</td>
                      <td className="px-4 py-3 text-sm text-gray-700 border-b">{formatDateTime(alert.create_time)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            {secretLabelAlerts.length === 0 && (
              <div className="text-center py-8 text-gray-500">
                暂无密标告警数据
              </div>
            )}
          </div>
        </div>
      );
    };

    const AlertOverviewContent = () => {
      const attackAlerts = allAlertsData?.data?.attack_alerts || [];
      const secretAlerts = allAlertsData?.data?.secret_alerts || [];
      const [selectedEvent, setSelectedEvent] = useState<any>(null);

      const getRiskLevel = (event: any) => {
        return '高危';
      };

      const getRiskLevelClass = (level: string) => {
        switch (level) {
          case '高危':
            return 'text-red-600';
          case '中危':
            return 'text-orange-600';
          case '低危':
            return 'text-yellow-600';
          default:
            return 'text-gray-600';
        }
      };

      const getAttackBehavior = (event: any) => {
        const type = event.type?.toLowerCase() || '';
        const info = event.info || '';
        
        if (type.includes('kubelet')) {
          if (info.includes('pod') || info.includes('pods')) {
            return '获取Pod列表';
          } else if (info.includes('container') || info.includes('containers')) {
            return '获取容器信息';
          } else if (info.includes('node') || info.includes('nodes')) {
            return '获取节点信息';
          } else if (info.includes('logs')) {
            return '查看日志';
          } else {
            return 'Kubelet API访问';
          }
        } else if (type.includes('docker')) {
          if (info.includes('ps') || info.includes('list')) {
            return '列出容器';
          } else if (info.includes('info')) {
            return '获取Docker信息';
          } else if (info.includes('exec')) {
            return '执行容器命令';
          } else {
            return 'Docker API访问';
          }
        } else if (type.includes('etcd')) {
          if (info.includes('get') || info.includes('read')) {
            return '读取配置数据';
          } else if (info.includes('put') || info.includes('write')) {
            return '写入配置数据';
          } else {
            return 'Etcd数据访问';
          }
        } else if (type.includes('apiserver')) {
          if (info.includes('pod') || info.includes('pods')) {
            return '访问Pod资源';
          } else if (info.includes('secret')) {
            return '访问Secret';
          } else if (info.includes('configmap')) {
            return '访问ConfigMap';
          } else {
            return 'Apiserver API访问';
          }
        } else if (type.includes('bash')) {
          if (info.includes('rm') || info.includes('delete')) {
            return '执行删除命令';
          } else if (info.includes('cat') || info.includes('read')) {
            return '读取文件';
          } else if (info.includes('wget') || info.includes('curl')) {
            return '下载文件';
          } else {
            return '执行Shell命令';
          }
        } else if (type.includes('ssh')) {
          return 'SSH登录尝试';
        } else if (type.includes('redis')) {
          return 'Redis访问';
        } else if (type.includes('mysql')) {
          return 'MySQL访问';
        } else if (type.includes('ftp')) {
          return 'FTP访问';
        } else if (type.includes('http')) {
          return 'HTTP访问';
        } else {
          return '未知攻击行为';
        }
      };

      return (
        <div className="space-y-6">
          {allAlertsError && (
            <div className="mb-4 p-4 bg-yellow-100 border border-yellow-400 text-yellow-700 rounded">
              ⚠️ 无法连接到后端服务器，请确保KubePot后端服务正在运行（端口9001）
            </div>
          )}

          {isLoadingAllAlerts && (
            <div className="flex items-center justify-center py-8">
              <div className="text-gray-600">加载中...</div>
            </div>
          )}

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="stat-card">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-gray-500 text-sm mb-1">攻击告警</p>
                  <p className="text-3xl font-bold text-red-600">{attackAlerts.length}</p>
                </div>
                <div className="w-12 h-12 bg-red-100 rounded-full flex items-center justify-center">
                  <svg className="w-6 h-6 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                  </svg>
                </div>
              </div>
            </div>

            <div className="stat-card">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-gray-500 text-sm mb-1">密标告警</p>
                  <p className="text-3xl font-bold text-orange-600">{secretAlerts.length}</p>
                </div>
                <div className="w-12 h-12 bg-orange-100 rounded-full flex items-center justify-center">
                  <svg className="w-6 h-6 text-orange-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                  </svg>
                </div>
              </div>
            </div>
          </div>

          <div className="bg-white rounded-lg p-6 shadow-sm">
            <div className="flex justify-between items-center mb-4">
              <h3 className="text-lg font-semibold text-gray-800">攻击告警列表</h3>
              <div className="flex items-center gap-2">
                <span className="text-sm text-gray-500">自动刷新：30秒</span>
                <button
                  onClick={() => refetchAllAlerts()}
                  className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600"
                >
                  刷新
                </button>
              </div>
            </div>
            <div className="overflow-x-auto">
              <table className="w-full border-collapse">
                <thead>
                  <tr className="bg-gray-50">
                    <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700 border-b">事件类型</th>
                    <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700 border-b">被攻击主机名</th>
                    <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700 border-b">被攻击IP</th>
                    <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700 border-b">被攻击节点/容器</th>
                    <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700 border-b">攻击来源</th>
                    <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700 border-b">攻击行为</th>
                    <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700 border-b">攻击时间</th>
                    <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700 border-b">风险等级</th>
                    <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700 border-b">操作</th>
                  </tr>
                </thead>
                <tbody>
                  {attackAlerts.map((event: any, index) => {
                    const riskLevel = getRiskLevel(event);
                    const getField = (obj: any, field: string) => {
                      return obj[field] || obj[field.charAt(0).toLowerCase() + field.slice(1)] || obj[field.toLowerCase()] || obj[field.replace(/([A-Z])/g, '_$1').toLowerCase()] || '-';
                    };
                    return (
                      <tr key={event.id} className={index % 2 === 0 ? 'bg-white' : 'bg-gray-50'}>
                        <td className="px-4 py-3 text-sm text-gray-700 border-b">{event.type}</td>
                        <td className="px-4 py-3 text-sm text-gray-700 border-b">{event.hostname || event.hostName || '-'}</td>
                        <td className="px-4 py-3 text-sm text-gray-700 border-b">{event.ip || '-'}</td>
                        <td className="px-4 py-3 text-sm text-gray-700 border-b">{event.nodetype || event.nodeType || '-'}</td>
                        <td className="px-4 py-3 text-sm text-gray-700 border-b">{event.agentip || event.agentIp || '-'}</td>
                        <td className="px-4 py-3 text-sm text-gray-700 border-b">{getAttackBehavior(event)}</td>
                        <td className="px-4 py-3 text-sm text-gray-700 border-b">{formatDateTime(event.create_time)}</td>
                        <td className={`px-4 py-3 text-sm font-semibold border-b ${getRiskLevelClass(riskLevel)}`}>{riskLevel}</td>
                        <td className="px-4 py-3 text-sm border-b">
                          <button
                            onClick={() => setSelectedEvent(event)}
                            className="text-blue-600 hover:text-blue-800 font-medium"
                          >
                            事件详情
                          </button>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
            {attackAlerts.length === 0 && (
              <div className="text-center py-8 text-gray-500">
                暂无攻击告警数据
              </div>
            )}
          </div>

          <div className="bg-white rounded-lg p-6 shadow-sm">
            <div className="flex justify-between items-center mb-4">
              <h3 className="text-lg font-semibold text-gray-800">密标告警列表</h3>
            </div>
            <div className="overflow-x-auto">
              <table className="w-full border-collapse">
                <thead>
                  <tr className="bg-gray-50">
                    <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700 border-b">密标名称</th>
                    <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700 border-b">Agent</th>
                    <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700 border-b">IP地址</th>
                    <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700 border-b">访问时间</th>
                    <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700 border-b">访问内容</th>
                    <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700 border-b">告警时间</th>
                  </tr>
                </thead>
                <tbody>
                  {secretAlerts.map((alert: any, index) => (
                    <tr key={alert.id} className={index % 2 === 0 ? 'bg-white' : 'bg-gray-50'}>
                      <td className="px-4 py-3 text-sm text-gray-700 border-b">{alert.secret_label_name}</td>
                      <td className="px-4 py-3 text-sm text-gray-700 border-b">{alert.agent}</td>
                      <td className="px-4 py-3 text-sm text-gray-700 border-b">{alert.ip}</td>
                      <td className="px-4 py-3 text-sm text-gray-700 border-b">{formatDateTime(alert.access_time)}</td>
                      <td className="px-4 py-3 text-sm text-gray-700 border-b max-w-xs truncate" title={alert.access_content}>{alert.access_content}</td>
                      <td className="px-4 py-3 text-sm text-gray-700 border-b">{formatDateTime(alert.create_time)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            {secretAlerts.length === 0 && (
              <div className="text-center py-8 text-gray-500">
                暂无密标告警数据
              </div>
            )}
          </div>

          {selectedEvent && (
            <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
              <div className="bg-white rounded-lg p-6 max-w-2xl w-full mx-4 shadow-xl">
                <div className="flex justify-between items-center mb-4">
                  <h3 className="text-lg font-semibold text-gray-800">事件详情</h3>
                  <button
                    onClick={() => setSelectedEvent(null)}
                    className="text-gray-500 hover:text-gray-700"
                  >
                    <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                    </svg>
                  </button>
                </div>
                <div className="space-y-3">
                  <div className="flex">
                    <span className="w-32 text-sm font-medium text-gray-600">事件类型:</span>
                    <span className="text-sm text-gray-800">{selectedEvent.type}</span>
                  </div>
                  <div className="flex">
                    <span className="w-32 text-sm font-medium text-gray-600">风险等级:</span>
                    <span className={`text-sm font-semibold ${getRiskLevelClass(getRiskLevel(selectedEvent))}`}>{getRiskLevel(selectedEvent)}</span>
                  </div>
                  <div className="flex">
                    <span className="w-32 text-sm font-medium text-gray-600">攻击行为:</span>
                    <span className="text-sm text-gray-800">{getAttackBehavior(selectedEvent)}</span>
                  </div>
                  <div className="flex">
                    <span className="w-32 text-sm font-medium text-gray-600">被攻击主机名:</span>
                    <span className="text-sm text-gray-800">{selectedEvent.hostname || selectedEvent.hostName || '-'}</span>
                  </div>
                  <div className="flex">
                    <span className="w-32 text-sm font-medium text-gray-600">被攻击IP:</span>
                    <span className="text-sm text-gray-800">{selectedEvent.ip || '-'}</span>
                  </div>
                  <div className="flex">
                    <span className="w-32 text-sm font-medium text-gray-600">被攻击节点:</span>
                    <span className="text-sm text-gray-800">{selectedEvent.nodetype || selectedEvent.nodeType || '-'}</span>
                  </div>
                  <div className="flex">
                    <span className="w-32 text-sm font-medium text-gray-600">攻击来源:</span>
                    <span className="text-sm text-gray-800">{selectedEvent.agentip || selectedEvent.agentIp || '-'}</span>
                  </div>
                  <div className="flex">
                    <span className="w-32 text-sm font-medium text-gray-600">额外信息:</span>
                    <span className="text-sm text-gray-800 break-all">{selectedEvent.info}</span>
                  </div>
                  <div className="flex">
                    <span className="w-32 text-sm font-medium text-gray-600">攻击时间:</span>
                    <span className="text-sm text-gray-800">{formatDateTime(selectedEvent.create_time)}</span>
                  </div>
                </div>
              </div>
            </div>
          )}
        </div>
      );
    };

    const envItems = [
      { key: 'node', label: '节点管理' },
      { key: 'node-install', label: '节点安装' }
    ];

    const AttackListContent = () => {
      const riskStats = getRiskStats();
      
      return (
      <div className="space-y-6">
        {attackEventsError && (
          <div className="mb-4 p-4 bg-yellow-100 border border-yellow-400 text-yellow-700 rounded">
            ⚠️ 无法连接到后端服务器，请确保KubePot后端服务正在运行（端口9001）
          </div>
        )}

        {isLoadingAttackEvents && (
          <div className="flex items-center justify-center py-8">
            <div className="text-gray-600">加载中...</div>
          </div>
        )}

        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          <div className="stat-card">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-gray-500 text-sm mb-1">高危事件</p>
                <p className="text-3xl font-bold text-red-600">{riskStats.high}</p>
              </div>
              <div className="w-12 h-12 bg-red-100 rounded-full flex items-center justify-center">
                <svg className="w-6 h-6 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                </svg>
              </div>
            </div>
          </div>

          <div className="stat-card">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-gray-500 text-sm mb-1">中危事件</p>
                <p className="text-3xl font-bold text-orange-600">{riskStats.medium}</p>
              </div>
              <div className="w-12 h-12 bg-orange-100 rounded-full flex items-center justify-center">
                <svg className="w-6 h-6 text-orange-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
              </div>
            </div>
          </div>

          <div className="stat-card">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-gray-500 text-sm mb-1">低危事件</p>
                <p className="text-3xl font-bold text-yellow-600">{riskStats.low}</p>
              </div>
              <div className="w-12 h-12 bg-yellow-100 rounded-full flex items-center justify-center">
                <svg className="w-6 h-6 text-yellow-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
              </div>
            </div>
          </div>
        </div>

        <div className="bg-white rounded-lg p-6 shadow-sm">
          <div className="flex justify-between items-center mb-4">
            <h3 className="text-lg font-semibold text-gray-800">事件列表</h3>
            <div className="flex items-center gap-2">
              <span className="text-sm text-gray-500">自动刷新：30秒</span>
              <button
                onClick={() => refetchAttackEvents()}
                className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600"
              >
                刷新
              </button>
            </div>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full border-collapse">
              <thead>
                <tr className="bg-gray-50">
                  <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700 border-b">事件类型</th>
                  <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700 border-b">被攻击主机名</th>
                  <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700 border-b">被攻击IP</th>
                  <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700 border-b">被攻击节点/容器</th>
                  <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700 border-b">攻击来源</th>
                  <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700 border-b">攻击行为</th>
                  <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700 border-b">攻击时间</th>
                  <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700 border-b">风险等级</th>
                  <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700 border-b">操作</th>
                </tr>
              </thead>
              <tbody>
                {attackEvents.map((event, index) => {
                  const riskLevel = getRiskLevel(event);
                  return (
                    <tr key={event.id} className={index % 2 === 0 ? 'bg-white' : 'bg-gray-50'}>
                      <td className="px-4 py-3 text-sm text-gray-700 border-b">{event.type}</td>
                      <td className="px-4 py-3 text-sm text-gray-700 border-b">{event.hostname || event.hostName || '-'}</td>
                      <td className="px-4 py-3 text-sm text-gray-700 border-b">{event.ip || '-'}</td>
                      <td className="px-4 py-3 text-sm text-gray-700 border-b">{event.nodetype || event.nodeType || '-'}</td>
                      <td className="px-4 py-3 text-sm text-gray-700 border-b">{event.agentip || event.agentIp || '-'}</td>
                      <td className="px-4 py-3 text-sm text-gray-700 border-b">{getAttackBehavior(event)}</td>
                      <td className="px-4 py-3 text-sm text-gray-700 border-b">{formatDateTime(event.create_time)}</td>
                      <td className={`px-4 py-3 text-sm font-semibold border-b ${getRiskLevelClass(riskLevel)}`}>{riskLevel}</td>
                      <td className="px-4 py-3 text-sm border-b">
                        <button
                          onClick={() => setSelectedEvent(event)}
                          className="text-blue-600 hover:text-blue-800 font-medium"
                        >
                          事件详情
                        </button>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
          <div className="mt-4 flex items-center justify-between">
            <div className="text-sm text-gray-500">
              第 {(currentPage - 1) * pageSize + 1}-{Math.min(currentPage * pageSize, parseInt(totalCount))} 条/总共 {totalCount} 条
            </div>
            <div className="flex items-center gap-2">
              <button
                onClick={() => setCurrentPage(p => Math.max(1, p - 1))}
                disabled={currentPage <= 1}
                className="px-3 py-1 border border-gray-300 rounded hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                上一页
              </button>
              <span className="px-2 text-sm text-gray-700">第 {currentPage} 页</span>
              <button
                onClick={() => setCurrentPage(p => p + 1)}
                disabled={currentPage * pageSize >= parseInt(totalCount)}
                className="px-3 py-1 border border-gray-300 rounded hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                下一页
              </button>
              <div className="flex items-center gap-1 ml-4">
                <span className="text-sm text-gray-700">跳转到</span>
                <input
                  type="number"
                  min={1}
                  max={Math.ceil(parseInt(totalCount) / pageSize) || 1}
                  value={jumpPage}
                  onChange={(e) => setJumpPage(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter') {
                      const page = parseInt(jumpPage);
                      const totalPages = Math.ceil(parseInt(totalCount) / pageSize);
                      if (!isNaN(page) && page >= 1 && page <= totalPages) {
                        setCurrentPage(page);
                        setJumpPage('');
                      }
                    }
                  }}
                  className="w-16 px-2 py-1 border border-gray-300 rounded text-center"
                  placeholder="页码"
                />
                <button
                  onClick={() => {
                    const page = parseInt(jumpPage);
                    const totalPages = Math.ceil(parseInt(totalCount) / pageSize);
                    if (!isNaN(page) && page >= 1 && page <= totalPages) {
                      setCurrentPage(page);
                      setJumpPage('');
                    }
                  }}
                  className="px-3 py-1 border border-gray-300 rounded hover:bg-gray-50"
                >
                  跳转
                </button>
              </div>
            </div>
          </div>
        </div>

        {selectedEvent && (
          <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
            <div className="bg-white rounded-lg p-6 max-w-2xl w-full mx-4 shadow-xl">
              <div className="flex justify-between items-center mb-4">
                <h3 className="text-lg font-semibold text-gray-800">事件详情</h3>
                <button
                  onClick={() => setSelectedEvent(null)}
                  className="text-gray-500 hover:text-gray-700"
                >
                  <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>
              <div className="space-y-3">
                <div className="flex">
                  <span className="w-32 text-sm font-medium text-gray-600">事件类型:</span>
                  <span className="text-sm text-gray-800">{selectedEvent.type}</span>
                </div>
                <div className="flex">
                  <span className="w-32 text-sm font-medium text-gray-600">风险等级:</span>
                  <span className={`text-sm font-semibold ${getRiskLevelClass(getRiskLevel(selectedEvent))}`}>{getRiskLevel(selectedEvent)}</span>
                </div>
                <div className="flex">
                  <span className="w-32 text-sm font-medium text-gray-600">攻击行为:</span>
                  <span className="text-sm text-gray-800">{getAttackBehavior(selectedEvent)}</span>
                </div>
                <div className="flex">
                  <span className="w-32 text-sm font-medium text-gray-600">被攻击主机名:</span>
                  <span className="text-sm text-gray-800">{selectedEvent.hostname}</span>
                </div>
                <div className="flex">
                  <span className="w-32 text-sm font-medium text-gray-600">被攻击IP:</span>
                  <span className="text-sm text-gray-800">{selectedEvent.ip}</span>
                </div>
                <div className="flex">
                  <span className="w-32 text-sm font-medium text-gray-600">被攻击节点:</span>
                  <span className="text-sm text-gray-800">{selectedEvent.nodetype}</span>
                </div>
                <div className="flex">
                  <span className="w-32 text-sm font-medium text-gray-600">攻击来源:</span>
                  <span className="text-sm text-gray-800">{selectedEvent.agentip}</span>
                </div>
                <div className="flex">
                  <span className="w-32 text-sm font-medium text-gray-600">额外信息:</span>
                  <span className="text-sm text-gray-800 break-all">{selectedEvent.info}</span>
                </div>
                <div className="flex">
                  <span className="w-32 text-sm font-medium text-gray-600">攻击时间:</span>
                  <span className="text-sm text-gray-800">{formatDateTime(selectedEvent.create_time)}</span>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>
      );
    };

    const NodeManagementContent = () => {
      const [showEditModal, setShowEditModal] = useState(false);
      const [editingColony, setEditingColony] = useState<any>(null);
      const [configData, setConfigData] = useState({
        ssh: 0,
        ftp: 0,
        http: 0,
        redis: 0,
        mysql: 0,
        telnet: 0,
        tftp: 0,
        vnc: 0,
        memCahe: 0,
        web: 0,
        es: 0,
        kubelet: 0,
        etcd: 0,
        apiserver: 0,
        docker: 0,
        bash: 0
      });
      const [isLoadingConfig, setIsLoadingConfig] = useState(false);
      const [isSavingConfig, setIsSavingConfig] = useState(false);

      // 计算时间差，返回多少分钟前
      const getMinutesAgo = (timeString: string) => {
        const lastUpdate = new Date(timeString);
        const now = new Date();
        const diffMs = now.getTime() - lastUpdate.getTime();
        const diffMins = Math.round(diffMs / 60000);
        return diffMins;
      };

      // 判断节点是否在线（基于最后更新时间，超过5分钟视为离线）
      const isNodeOnline = (lastUpdateTime: string) => {
        const minsAgo = getMinutesAgo(lastUpdateTime);
        return minsAgo <= 5;
      };

      const formatDateTime = (dateTime: string) => {
        if (!dateTime) return '-';
        return dateTime.split('+')[0].split('T').join(' ');
      };
      
      // 格式化最后更新时间显示
      const formatLastUpdateTime = (timeString: string) => {
        const minsAgo = getMinutesAgo(timeString);
        if (minsAgo < 1) {
          return '刚刚';
        } else if (minsAgo < 60) {
          return `${minsAgo}分钟前`;
        } else if (minsAgo < 1440) {
          const hours = Math.floor(minsAgo / 60);
          return `${hours}小时前`;
        } else {
          const days = Math.floor(minsAgo / 1440);
          return `${days}天前`;
        }
      };

      const openEditModal = async (colony: any) => {
        setEditingColony(colony);
        setIsLoadingConfig(true);
        try {
          const response = await dashboardApi.getAgentConfig(colony.agent_name);
          console.log('获取节点配置响应:', response);
          if (response.code === 0 || response.code === 200) {
            console.log('后端返回的配置数据:', response.data);
            const data = response.data;
            setConfigData({
              ssh: Number(data.ssh) || 0,
              ftp: Number(data.ftp) || 0,
              http: Number(data.http) || 0,
              redis: Number(data.redis) || 0,
              mysql: Number(data.mysql) || 0,
              telnet: Number(data.telnet) || 0,
              tftp: Number(data.tftp) || 0,
              vnc: Number(data.vnc) || 0,
              memCahe: Number(data.mem_cahe) || 0,
              web: Number(data.web) || 0,
              es: Number(data.es) || 0,
              kubelet: Number(data.kubelet) || 0,
              etcd: Number(data.etcd) || 0,
              apiserver: Number(data.apiserver) || 0,
              docker: Number(data.docker) || 0,
              bash: Number(data.bash) || 0
            });
            console.log('设置后的 configData:', {
              ssh: Number(data.ssh) || 0,
              ftp: Number(data.ftp) || 0,
              http: Number(data.http) || 0,
              redis: Number(data.redis) || 0,
              mysql: Number(data.mysql) || 0,
              telnet: Number(data.telnet) || 0,
              tftp: Number(data.tftp) || 0,
              vnc: Number(data.vnc) || 0,
              memCahe: Number(data.mem_cahe) || 0,
              web: Number(data.web) || 0,
              es: Number(data.es) || 0,
              kubelet: Number(data.kubelet) || 0,
              etcd: Number(data.etcd) || 0,
              apiserver: Number(data.apiserver) || 0,
              docker: Number(data.docker) || 0,
              bash: Number(data.bash) || 0
            });
          } else {
            // 如果获取失败，使用默认值
            setConfigData({
              ssh: 0,
              ftp: 0,
              http: 0,
              redis: 0,
              mysql: 0,
              telnet: 0,
              tftp: 0,
              vnc: 0,
              memCahe: 0,
              web: 0,
              es: 0,
              kubelet: 0,
              etcd: 0,
              apiserver: 0,
              docker: 0,
              bash: 0
            });
          }
        } catch (error) {
          console.error('获取节点配置失败:', error);
          setConfigData({
            ssh: 0,
            ftp: 0,
            http: 0,
            redis: 0,
            mysql: 0,
            telnet: 0,
            tftp: 0,
            vnc: 0,
            memCahe: 0,
            web: 0,
            es: 0,
            kubelet: 0,
            etcd: 0,
            apiserver: 0,
            docker: 0,
            bash: 0
          });
        } finally {
          setIsLoadingConfig(false);
          setShowEditModal(true);
        }
      };

      const handleSaveConfig = async () => {
        if (!editingColony) return;
        
        setIsSavingConfig(true);
        try {
          const response = await dashboardApi.updateAgentConfig({
            agent_name: editingColony.agent_name,
            ssh: String(configData.ssh),
            ftp: String(configData.ftp),
            http: String(configData.http),
            redis: String(configData.redis),
            mysql: String(configData.mysql),
            telnet: String(configData.telnet),
            tftp: String(configData.tftp),
            vnc: String(configData.vnc),
            mem_cahe: String(configData.memCahe),
            web: String(configData.web || 0),
            es: String(configData.es || 0),
            kubelet: String(configData.kubelet || 0),
            etcd: String(configData.etcd || 0),
            apiserver: String(configData.apiserver || 0),
            docker: String(configData.docker || 0),
            bash: String(configData.bash || 0)
          });
          
          if (response.code === 0 || response.code === 200) {
            setShowEditModal(false);
            refetchColony();
            message.success('配置保存成功');
          } else {
            message.error('配置保存失败: ' + response.msg);
          }
        } catch (error) {
          console.error('保存节点配置失败:', error);
          message.error('配置保存失败，请检查网络连接');
        } finally {
          setIsSavingConfig(false);
        }
      };

      return (
        <div className="space-y-6">
          {colonyError && (
            <div className="mb-4 p-4 bg-yellow-100 border border-yellow-400 text-yellow-700 rounded">
              ⚠️ 无法连接到后端服务器，请确保KubePot后端服务正在运行（端口9001）
            </div>
          )}

          {isLoadingColony && (
            <div className="flex items-center justify-center py-8">
              <div className="text-gray-600">加载中...</div>
            </div>
          )}

          <div className="bg-white rounded-lg p-6 shadow-sm">
            <div className="flex justify-between items-center mb-4">
              <h3 className="text-lg font-semibold text-gray-800">节点列表</h3>
              <button
                onClick={() => refetchColony()}
                className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600"
              >
                刷新
              </button>
            </div>
            <div className="overflow-x-auto">
              <table className="w-full border-collapse">
                <thead>
                  <tr className="bg-gray-50">
                    <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700 border-b">节点名称</th>
                    <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700 border-b">节点地址</th>
                    <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700 border-b">状态</th>
                    <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700 border-b">最后更新时间</th>
                    <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700 border-b">操作</th>
                  </tr>
                </thead>
                <tbody>
                  {colonies.map((colony, index) => (
                    <tr key={colony.id} className={index % 2 === 0 ? 'bg-white' : 'bg-gray-50'}>
                      <td className="px-4 py-3 text-sm text-gray-700 border-b">{colony.agent_name}</td>
                      <td className="px-4 py-3 text-sm text-gray-700 border-b">{colony.agent_ip || '-'}</td>
                      <td className="px-4 py-3 text-sm border-b">
                        <span className={`px-3 py-1 rounded-full text-xs font-medium ${
                          isNodeOnline(colony.last_update_time)
                            ? 'bg-green-100 text-green-800' 
                            : 'bg-red-100 text-red-800'
                        }`}>
                          {isNodeOnline(colony.last_update_time) ? '在线' : '离线'}
                        </span>
                      </td>
                      <td className="px-4 py-3 text-sm text-gray-700 border-b">
                        <div>
                          <div>{formatDateTime(colony.last_update_time)}</div>
                        <div className="text-xs text-gray-500">{formatLastUpdateTime(colony.last_update_time)}</div>
                        </div>
                      </td>
                      <td className="px-4 py-3 text-sm border-b space-x-2">
                        <button
                          onClick={() => openEditModal(colony)}
                          className="text-blue-600 hover:text-blue-800 font-medium"
                        >
                          编辑
                        </button>
                        <button
                          onClick={() => handleDeleteColony(colony.id)}
                          className="text-red-600 hover:text-red-800 font-medium"
                        >
                          删除
                        </button>
                        <button
                          onClick={() => handleUninstallAgent(colony.agent_name, colony.agent_ip)}
                          className="text-orange-600 hover:text-orange-800 font-medium"
                        >
                          卸载
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            {colonies.length === 0 && (
              <div className="text-center py-8 text-gray-500">
                暂无节点数据
              </div>
            )}
          </div>

          {showEditModal && editingColony && (
            <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
              <div className="bg-white rounded-lg p-6 max-w-2xl w-full mx-4 shadow-xl">
                <div className="flex justify-between items-center mb-4">
                  <h3 className="text-lg font-semibold text-gray-800">编辑节点配置 - {editingColony.agent_name}</h3>
                  <button
                    onClick={() => setShowEditModal(false)}
                    className="text-gray-500 hover:text-gray-700"
                  >
                    <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                    </svg>
                  </button>
                </div>
                
                {isLoadingConfig ? (
                  <div className="flex items-center justify-center py-8">
                    <div className="text-gray-600">加载配置中...</div>
                  </div>
                ) : (
                  <div className="space-y-4">
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                      <div className="flex items-center justify-between p-3 border border-gray-200 rounded">
                        <span className="font-medium">WEB服务</span>
                        <label className="relative inline-flex items-center cursor-pointer">
                          <input 
                            type="checkbox" 
                            className="sr-only peer"
                            checked={configData.web == 1}
                            onChange={(e) => setConfigData({ ...configData, web: e.target.checked ? 1 : 0 })}
                          />
                          <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-500"></div>
                        </label>
                      </div>
                      
                      <div className="flex items-center justify-between p-3 border border-gray-200 rounded">
                        <span className="font-medium">SSH服务</span>
                        <label className="relative inline-flex items-center cursor-pointer">
                          <input 
                            type="checkbox" 
                            className="sr-only peer"
                            checked={configData.ssh == 1}
                            onChange={(e) => setConfigData({ ...configData, ssh: e.target.checked ? 1 : 0 })}
                          />
                          <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-500"></div>
                        </label>
                      </div>
                      
                      <div className="flex items-center justify-between p-3 border border-gray-200 rounded">
                        <span className="font-medium">FTP服务</span>
                        <label className="relative inline-flex items-center cursor-pointer">
                          <input 
                            type="checkbox" 
                            className="sr-only peer"
                            checked={configData.ftp == 1}
                            onChange={(e) => setConfigData({ ...configData, ftp: e.target.checked ? 1 : 0 })}
                          />
                          <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-500"></div>
                        </label>
                      </div>
                      
                      <div className="flex items-center justify-between p-3 border border-gray-200 rounded">
                        <span className="font-medium">HTTP服务</span>
                        <label className="relative inline-flex items-center cursor-pointer">
                          <input 
                            type="checkbox" 
                            className="sr-only peer"
                            checked={configData.http == 1}
                            onChange={(e) => setConfigData({ ...configData, http: e.target.checked ? 1 : 0 })}
                          />
                          <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-500"></div>
                        </label>
                      </div>
                      
                      <div className="flex items-center justify-between p-3 border border-gray-200 rounded">
                        <span className="font-medium">Redis服务</span>
                        <label className="relative inline-flex items-center cursor-pointer">
                          <input 
                            type="checkbox" 
                            className="sr-only peer"
                            checked={configData.redis == 1}
                            onChange={(e) => setConfigData({ ...configData, redis: e.target.checked ? 1 : 0 })}
                          />
                          <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-500"></div>
                        </label>
                      </div>
                      
                      <div className="flex items-center justify-between p-3 border border-gray-200 rounded">
                        <span className="font-medium">MySQL服务</span>
                        <label className="relative inline-flex items-center cursor-pointer">
                          <input 
                            type="checkbox" 
                            className="sr-only peer"
                            checked={configData.mysql == 1}
                            onChange={(e) => setConfigData({ ...configData, mysql: e.target.checked ? 1 : 0 })}
                          />
                          <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-500"></div>
                        </label>
                      </div>
                      
                      <div className="flex items-center justify-between p-3 border border-gray-200 rounded">
                        <span className="font-medium">Telnet服务</span>
                        <label className="relative inline-flex items-center cursor-pointer">
                          <input 
                            type="checkbox" 
                            className="sr-only peer"
                            checked={configData.telnet == 1}
                            onChange={(e) => setConfigData({ ...configData, telnet: e.target.checked ? 1 : 0 })}
                          />
                          <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-500"></div>
                        </label>
                      </div>
                      
                      <div className="flex items-center justify-between p-3 border border-gray-200 rounded">
                        <span className="font-medium">TFTP服务</span>
                        <label className="relative inline-flex items-center cursor-pointer">
                          <input 
                            type="checkbox" 
                            className="sr-only peer"
                            checked={configData.tftp == 1}
                            onChange={(e) => setConfigData({ ...configData, tftp: e.target.checked ? 1 : 0 })}
                          />
                          <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-500"></div>
                        </label>
                      </div>
                      
                      <div className="flex items-center justify-between p-3 border border-gray-200 rounded">
                        <span className="font-medium">VNC服务</span>
                        <label className="relative inline-flex items-center cursor-pointer">
                          <input 
                            type="checkbox" 
                            className="sr-only peer"
                            checked={configData.vnc == 1}
                            onChange={(e) => setConfigData({ ...configData, vnc: e.target.checked ? 1 : 0 })}
                          />
                          <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-500"></div>
                        </label>
                      </div>
                      
                      <div className="flex items-center justify-between p-3 border border-gray-200 rounded">
                        <span className="font-medium">MemCache服务</span>
                        <label className="relative inline-flex items-center cursor-pointer">
                          <input 
                            type="checkbox" 
                            className="sr-only peer"
                            checked={configData.memCahe == 1}
                            onChange={(e) => setConfigData({ ...configData, memCahe: e.target.checked ? 1 : 0 })}
                          />
                          <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-500"></div>
                        </label>
                      </div>
                      
                      <div className="flex items-center justify-between p-3 border border-gray-200 rounded">
                        <span className="font-medium">ES服务</span>
                        <label className="relative inline-flex items-center cursor-pointer">
                          <input 
                            type="checkbox" 
                            className="sr-only peer"
                            checked={configData.es == 1}
                            onChange={(e) => setConfigData({ ...configData, es: e.target.checked ? 1 : 0 })}
                          />
                          <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-500"></div>
                        </label>
                      </div>
                      
                      <div className="flex items-center justify-between p-3 border border-gray-200 rounded">
                        <span className="font-medium">Kubelet服务</span>
                        <label className="relative inline-flex items-center cursor-pointer">
                          <input 
                            type="checkbox" 
                            className="sr-only peer"
                            checked={configData.kubelet == 1}
                            onChange={(e) => setConfigData({ ...configData, kubelet: e.target.checked ? 1 : 0 })}
                          />
                          <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-500"></div>
                        </label>
                      </div>
                      
                      <div className="flex items-center justify-between p-3 border border-gray-200 rounded">
                        <span className="font-medium">Etcd服务</span>
                        <label className="relative inline-flex items-center cursor-pointer">
                          <input 
                            type="checkbox" 
                            className="sr-only peer"
                            checked={configData.etcd == 1}
                            onChange={(e) => setConfigData({ ...configData, etcd: e.target.checked ? 1 : 0 })}
                          />
                          <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-500"></div>
                        </label>
                      </div>
                      
                      <div className="flex items-center justify-between p-3 border border-gray-200 rounded">
                        <span className="font-medium">Apiserver服务</span>
                        <label className="relative inline-flex items-center cursor-pointer">
                          <input 
                            type="checkbox" 
                            className="sr-only peer"
                            checked={configData.apiserver == 1}
                            onChange={(e) => setConfigData({ ...configData, apiserver: e.target.checked ? 1 : 0 })}
                          />
                          <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-500"></div>
                        </label>
                      </div>
                      
                      <div className="flex items-center justify-between p-3 border border-gray-200 rounded">
                        <span className="font-medium">Docker服务</span>
                        <label className="relative inline-flex items-center cursor-pointer">
                          <input 
                            type="checkbox" 
                            className="sr-only peer"
                            checked={configData.docker == 1}
                            onChange={(e) => setConfigData({ ...configData, docker: e.target.checked ? 1 : 0 })}
                          />
                          <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-500"></div>
                        </label>
                      </div>
                      
                      <div className="flex items-center justify-between p-3 border border-gray-200 rounded">
                        <span className="font-medium">Bash服务</span>
                        <label className="relative inline-flex items-center cursor-pointer">
                          <input 
                            type="checkbox" 
                            className="sr-only peer"
                            checked={configData.bash == 1}
                            onChange={(e) => setConfigData({ ...configData, bash: e.target.checked ? 1 : 0 })}
                          />
                          <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-500"></div>
                        </label>
                      </div>
                    </div>
                    
                    <div className="flex justify-end gap-2">
                      <button
                        onClick={() => setShowEditModal(false)}
                        className="px-4 py-2 bg-gray-200 text-gray-800 rounded hover:bg-gray-300"
                      >
                        取消
                      </button>
                      <button
                        onClick={handleSaveConfig}
                        disabled={isSavingConfig}
                        className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed"
                      >
                        {isSavingConfig ? '保存中...' : '保存'}
                      </button>
                    </div>
                  </div>
                )}
              </div>
            </div>
          )}
        </div>
      );
    };

    const NodeInstallContent = () => {
      const serverUrl = window.location.origin;
      
      const copyCommand = (command: string) => {
        navigator.clipboard.writeText(command).then(() => {
          message.success('命令已复制到剪贴板！');
        }).catch(() => {
          message.error('复制失败，请手动复制');
        });
      };

      const packages = [
        {
          os: 'Linux (x86_64)',
          fileName: 'kubepot-agent-linux-amd64',
          command: `curl -o kubepot-agent ${serverUrl}/download/kubepot-agent-linux-amd64 && chmod +x kubepot-agent && ./kubepot-agent`,
          description: '适用于 Linux x86_64 系统'
        },
        {
          os: 'Linux (ARM64)',
          fileName: 'kubepot-agent-linux-arm64',
          command: `curl -o kubepot-agent ${serverUrl}/download/kubepot-agent-linux-arm64 && chmod +x kubepot-agent && ./kubepot-agent`,
          description: '适用于 Linux ARM64 系统（如树莓派、ARM服务器）'
        },
        {
          os: 'macOS (Intel)',
          fileName: 'kubepot-agent-darwin-amd64',
          command: `curl -o kubepot-agent ${serverUrl}/download/kubepot-agent-darwin-amd64 && chmod +x kubepot-agent && ./kubepot-agent`,
          description: '适用于 macOS Intel 芯片'
        },
        {
          os: 'macOS (Apple Silicon)',
          fileName: 'kubepot-agent-darwin-arm64',
          command: `curl -o kubepot-agent ${serverUrl}/download/kubepot-agent-darwin-arm64 && chmod +x kubepot-agent && ./kubepot-agent`,
          description: '适用于 macOS Apple Silicon (M1/M2/M3) 芯片'
        },
        {
          os: 'Windows (x86_64)',
          fileName: 'kubepot-agent-windows-amd64.exe',
          command: `Invoke-WebRequest -Uri "${serverUrl}/download/kubepot-agent-windows-amd64.exe" -OutFile "kubepot-agent.exe"; .\\kubepot-agent.exe`,
          description: '适用于 Windows x86_64 系统（使用 PowerShell）'
        }
      ];

      return (
        <div className="space-y-6">
          <div className="bg-white rounded-lg p-6 shadow-sm">
            <div className="mb-6">
              <h3 className="text-lg font-semibold text-gray-800 mb-2">节点安装</h3>
              <p className="text-gray-600">选择适合您系统的安装包，复制命令在目标节点上执行即可完成安装</p>
            </div>

            <div className="grid grid-cols-1 gap-6">
              {packages.map((pkg, index) => (
                <div key={index} className="border border-gray-200 rounded-lg p-4">
                  <div className="flex justify-between items-start mb-3">
                    <div>
                      <h4 className="font-semibold text-gray-800">{pkg.os}</h4>
                      <p className="text-sm text-gray-500 mt-1">{pkg.description}</p>
                    </div>
                    <button
                      onClick={() => copyCommand(pkg.command)}
                      className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 text-sm"
                    >
                      复制命令
                    </button>
                  </div>
                  <div className="bg-gray-900 text-gray-100 p-3 rounded font-mono text-sm overflow-x-auto">
                    <code>{pkg.command}</code>
                  </div>
                </div>
              ))}
            </div>

            <div className="mt-6 p-4 bg-blue-50 border border-blue-200 rounded-lg">
              <h4 className="font-semibold text-blue-800 mb-2">安装说明</h4>
              <ul className="text-sm text-blue-700 space-y-1">
                <li>• 确保目标节点可以访问服务端地址：<code className="bg-blue-100 px-1 rounded">{serverUrl}</code></li>
                <li>• 执行命令后，Agent 会自动连接到服务端</li>
                <li>• 安装成功后，可以在「节点管理」中看到新加入的节点</li>
                <li>• 如需后台运行，请使用 nohup 或 systemd 等方式管理进程</li>
              </ul>
            </div>
          </div>
        </div>
      );
    };

    const AttackerProfileContent = () => {
      const attackerStats = () => {
        const ipMap = new Map<string, { count: number; types: Set<string>; lastTime: string; country: string }>();
        
        attackEvents.forEach(event => {
          const ip = event.ip || event.agentip || '未知';
          if (!ipMap.has(ip)) {
            ipMap.set(ip, { count: 0, types: new Set(), lastTime: event.create_time, country: event.country || '未知' });
          }
          const stats = ipMap.get(ip)!;
          stats.count++;
          stats.types.add(event.type);
          if (event.create_time > stats.lastTime) {
            stats.lastTime = event.create_time;
          }
        });

        return Array.from(ipMap.entries())
          .map(([ip, stats]) => ({
            ip,
            count: stats.count,
            types: Array.from(stats.types),
            lastTime: stats.lastTime,
            country: stats.country
          }))
          .sort((a, b) => b.count - a.count);
      };

      const attackers = attackerStats();

      const getAttackerRiskLevel = (count: number) => {
        return { label: '高危', class: 'text-red-600 bg-red-100' };
      };

      return (
        <div className="space-y-6">
          {attackEventsError && (
            <div className="mb-4 p-4 bg-yellow-100 border border-yellow-400 text-yellow-700 rounded">
              ⚠️ 无法连接到后端服务器，请确保KubePot后端服务正在运行（端口9001）
            </div>
          )}

          {isLoadingAttackEvents && (
            <div className="flex items-center justify-center py-8">
              <div className="text-gray-600">加载中...</div>
            </div>
          )}

          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            <div className="stat-card">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-gray-500 text-sm mb-1">攻击者数量</p>
                  <p className="text-3xl font-bold text-blue-600">{attackers.length}</p>
                </div>
                <div className="w-12 h-12 bg-blue-100 rounded-full flex items-center justify-center">
                  <svg className="w-6 h-6 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
                  </svg>
                </div>
              </div>
            </div>

            <div className="stat-card">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-gray-500 text-sm mb-1">总攻击次数</p>
                  <p className="text-3xl font-bold text-red-600">{attackEvents.length}</p>
                </div>
                <div className="w-12 h-12 bg-red-100 rounded-full flex items-center justify-center">
                  <svg className="w-6 h-6 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
                  </svg>
                </div>
              </div>
            </div>

            <div className="stat-card">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-gray-500 text-sm mb-1">攻击类型数</p>
                  <p className="text-3xl font-bold text-purple-600">
                    {new Set(attackEvents.map(e => e.type)).size}
                  </p>
                </div>
                <div className="w-12 h-12 bg-purple-100 rounded-full flex items-center justify-center">
                  <svg className="w-6 h-6 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z" />
                  </svg>
                </div>
              </div>
            </div>
          </div>

          <div className="bg-white rounded-lg p-6 shadow-sm">
            <div className="flex justify-between items-center mb-4">
              <h3 className="text-lg font-semibold text-gray-800">攻击者列表</h3>
              <div className="flex items-center gap-2">
                <span className="text-sm text-gray-500">自动刷新：30秒</span>
                <button
                  onClick={() => refetchAttackEvents()}
                  className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600"
                >
                  刷新
                </button>
              </div>
            </div>
            <div className="overflow-x-auto">
              <table className="w-full border-collapse">
                <thead>
                  <tr className="bg-gray-50">
                    <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700 border-b">IP地址</th>
                    <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700 border-b">地理位置</th>
                    <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700 border-b">攻击次数</th>
                    <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700 border-b">攻击类型</th>
                    <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700 border-b">风险等级</th>
                    <th className="px-4 py-3 text-left text-sm font-semibold text-gray-700 border-b">最后攻击时间</th>
                  </tr>
                </thead>
                <tbody>
                  {attackers.map((attacker, index) => {
                    const riskLevel = getAttackerRiskLevel(attacker.count);
                    return (
                      <tr key={attacker.ip} className={index % 2 === 0 ? 'bg-white' : 'bg-gray-50'}>
                        <td className="px-4 py-3 text-sm text-gray-700 border-b font-mono">{attacker.ip}</td>
                        <td className="px-4 py-3 text-sm text-gray-700 border-b">{attacker.country}</td>
                        <td className="px-4 py-3 text-sm text-gray-700 border-b font-semibold">{attacker.count}</td>
                        <td className="px-4 py-3 text-sm text-gray-700 border-b">
                          <div className="flex flex-wrap gap-1">
                            {attacker.types.slice(0, 3).map((type, i) => (
                              <span key={i} className="px-2 py-0.5 bg-gray-100 text-gray-600 rounded text-xs">
                                {type}
                              </span>
                            ))}
                            {attacker.types.length > 3 && (
                              <span className="px-2 py-0.5 bg-gray-100 text-gray-600 rounded text-xs">
                                +{attacker.types.length - 3}
                              </span>
                            )}
                          </div>
                        </td>
                        <td className="px-4 py-3 text-sm border-b">
                          <span className={`px-3 py-1 rounded-full text-xs font-medium ${riskLevel.class}`}>
                            {riskLevel.label}
                          </span>
                        </td>
                        <td className="px-4 py-3 text-sm text-gray-700 border-b">{attacker.lastTime}</td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
            {attackers.length === 0 && (
              <div className="text-center py-8 text-gray-500">
                暂无攻击者数据
              </div>
            )}
          </div>
        </div>
      );
    };

    const CountermeasureConfigContent = () => {
      return (
        <div className="space-y-6">
          <div className="bg-white rounded-lg p-6 shadow-sm">
            <h3 className="text-lg font-semibold text-gray-800 mb-4">反制配置</h3>
            <p className="text-gray-600 mb-4">反制功能配置模块</p>
            
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="p-4 border border-gray-200 rounded-lg">
                <h4 className="font-semibold text-gray-800 mb-2">反制规则管理</h4>
                <p className="text-sm text-gray-600">管理反制规则配置</p>
              </div>
              <div className="p-4 border border-gray-200 rounded-lg">
                <h4 className="font-semibold text-gray-800 mb-2">反制记录</h4>
                <p className="text-sm text-gray-600">查看反制操作历史记录</p>
              </div>
            </div>
          </div>
        </div>
      );
    };

    const TraceabilityConfigContent = () => {
      return (
        <div className="space-y-6">
          <div className="bg-white rounded-lg p-6 shadow-sm">
            <h3 className="text-lg font-semibold text-gray-800 mb-4">溯源配置</h3>
            <p className="text-gray-600 mb-4">溯源功能配置模块</p>
            
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="p-4 border border-gray-200 rounded-lg">
                <h4 className="font-semibold text-gray-800 mb-2">溯源规则配置</h4>
                <p className="text-sm text-gray-600">配置溯源规则和策略</p>
              </div>
              <div className="p-4 border border-gray-200 rounded-lg">
                <h4 className="font-semibold text-gray-800 mb-2">溯源记录</h4>
                <p className="text-sm text-gray-600">查看溯源操作历史记录</p>
              </div>
            </div>
          </div>
        </div>
      );
    };

    return (
      <div className="flex h-full bg-white">
        <div className="w-[208px] flex-shrink-0 bg-white h-[calc(100vh-56px)]">
          <div className="flex flex-col gap-2 items-center justify-center p-4">
            <div className="w-full border-t pt-4">
              <div className="text-sm font-semibold mb-2 px-4">总览</div>
              {overviewItems.map(item => (
                <div
                  key={item.key}
                  onClick={() => setActiveMenu(item.key)}
                  className={`px-4 py-2 text-sm cursor-pointer hover:bg-gray-100 ${
                    activeMenu === item.key ? 'bg-[#EDF0F5] text-blue-500' : 'text-black'
                  }`}
                >
                  {item.label}
                </div>
              ))}
            </div>

            <div className="w-full border-t pt-4">
              <div className="text-sm font-semibold mb-2 px-4">威胁感知</div>
              {threatItems.map(item => (
                <div
                  key={item.key}
                  onClick={() => setActiveMenu(item.key)}
                  className={`px-4 py-2 text-sm cursor-pointer hover:bg-gray-100 ${
                    activeMenu === item.key ? 'bg-[#EDF0F5] text-blue-500' : 'text-black'
                  }`}
                >
                  {item.label}
                </div>
              ))}
            </div>

            <div className="w-full border-t pt-4">
              <div className="text-sm font-semibold mb-2 px-4">密标管理</div>
              {secretItems.map(item => (
                <div
                  key={item.key}
                  onClick={() => setActiveMenu(item.key)}
                  className={`px-4 py-2 text-sm cursor-pointer hover:bg-gray-100 ${
                    activeMenu === item.key ? 'bg-[#EDF0F5] text-blue-500' : 'text-black'
                  }`}
                >
                  {item.label}
                </div>
              ))}
            </div>

            <div className="w-full border-t pt-4">
              <div className="text-sm font-semibold mb-2 px-4">攻击者画像</div>
              {attackerItems.map(item => (
                <div
                  key={item.key}
                  onClick={() => setActiveMenu(item.key)}
                  className={`px-4 py-2 text-sm cursor-pointer hover:bg-gray-100 ${
                    activeMenu === item.key ? 'bg-[#EDF0F5] text-blue-500' : 'text-black'
                  }`}
                >
                  {item.label}
                </div>
              ))}
            </div>

            <div className="w-full border-t pt-4">
              <div className="text-sm font-semibold mb-2 px-4">反制与溯源</div>
              {countermeasureItems.map(item => (
                <div
                  key={item.key}
                  onClick={() => setActiveMenu(item.key)}
                  className={`px-4 py-2 text-sm cursor-pointer hover:bg-gray-100 ${
                    activeMenu === item.key ? 'bg-[#EDF0F5] text-blue-500' : 'text-black'
                  }`}
                >
                  {item.label}
                </div>
              ))}
            </div>

            <div className="w-full border-t pt-4">
              <div className="text-sm font-semibold mb-2 px-4">环境管理</div>
              {envItems.map(item => (
                <div
                  key={item.key}
                  onClick={() => setActiveMenu(item.key)}
                  className={`px-4 py-2 text-sm cursor-pointer hover:bg-gray-100 ${
                    activeMenu === item.key ? 'bg-[#EDF0F5] text-blue-500' : 'text-black'
                  }`}
                >
                  {item.label}
                </div>
              ))}
            </div>
          </div>
        </div>

        <div className="flex-grow bg-[#EEF0F6] rounded-lg p-6">
          <div className="h-full">
            {activeMenu === 'overview' ? (
              <MonitorPanel />
            ) : activeMenu === 'alert-overview' ? (
              <AlertOverviewContent />
            ) : activeMenu === 'attack-list' ? (
              <AttackListContent />
            ) : activeMenu === 'attacker-profile' ? (
              <AttackerProfileContent />
            ) : activeMenu === 'countermeasure-config' ? (
              <CountermeasureConfigContent />
            ) : activeMenu === 'traceability-config' ? (
              <TraceabilityConfigContent />
            ) : activeMenu === 'node' ? (
              <NodeManagementContent />
            ) : activeMenu === 'node-install' ? (
              <NodeInstallContent />
            ) : activeMenu === 'secret-label' ? (
              <SecretLabelContent />
            ) : activeMenu === 'secret-label-alert' ? (
              <SecretLabelAlertContent />
            ) : (
              <div className="bg-white rounded-lg p-6 shadow-sm">
                <h2 className="text-2xl font-bold text-gray-800 mb-4">
                  {overviewItems.find(item => item.key === activeMenu)?.label || 
                   threatItems.find(item => item.key === activeMenu)?.label || 
                   attackerItems.find(item => item.key === activeMenu)?.label || 
                   countermeasureItems.find(item => item.key === activeMenu)?.label || 
                   secretItems.find(item => item.key === activeMenu)?.label || 
                   envItems.find(item => item.key === activeMenu)?.label || '详情面板'}
                </h2>
                <p className="text-gray-600">内容区域</p>
              </div>
            )}
          </div>
        </div>
      </div>
    );
  };

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="bg-white shadow-sm border-b">
        <div className="container mx-auto px-6">
          <div className="flex space-x-8">
          </div>
        </div>
      </div>

      <div className="min-h-[calc(100vh-64px)]">
        <MainContent />
      </div>
    </div>
  );
};

export default Dashboard;
