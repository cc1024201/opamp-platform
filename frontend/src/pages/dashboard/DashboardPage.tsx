import { useEffect } from 'react';
import {
  Box,
  Paper,
  Typography,
  Card,
  CardContent,
} from '@mui/material';
import {
  DevicesOther as AgentsIcon,
  CheckCircle as ConnectedIcon,
  Error as DisconnectedIcon,
  Settings as ConfigIcon,
} from '@mui/icons-material';
import { PieChart, Pie, Cell, ResponsiveContainer, Legend, Tooltip } from 'recharts';
import { useAgentStore } from '@/stores/agentStore';
import { useConfigurationStore } from '@/stores/configurationStore';

interface StatCardProps {
  title: string;
  value: number;
  icon: React.ReactNode;
  color: string;
}

function StatCard({ title, value, icon, color }: StatCardProps) {
  return (
    <Card>
      <CardContent>
        <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
          <Box
            sx={{
              backgroundColor: color,
              borderRadius: '50%',
              width: 48,
              height: 48,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              mr: 2,
            }}
          >
            {icon}
          </Box>
          <Box>
            <Typography variant="h4" component="div">
              {value}
            </Typography>
            <Typography variant="body2" color="text.secondary">
              {title}
            </Typography>
          </Box>
        </Box>
      </CardContent>
    </Card>
  );
}

export default function DashboardPage() {
  const { agents, total, fetchAgents } = useAgentStore();
  const { configurations, fetchConfigurations } = useConfigurationStore();

  useEffect(() => {
    fetchAgents();
    fetchConfigurations();
  }, [fetchAgents, fetchConfigurations]);

  const connectedAgents = agents.filter((a) => a.status === 'connected').length;
  const disconnectedAgents = agents.filter(
    (a) => a.status === 'disconnected' || a.status === 'error'
  ).length;
  const configuringAgents = agents.filter((a) => a.status === 'configuring').length;

  // Agent 状态分布数据
  const statusData = [
    { name: '在线', value: connectedAgents, color: '#2e7d32' },
    { name: '离线', value: disconnectedAgents, color: '#d32f2f' },
    { name: '配置中', value: configuringAgents, color: '#ed6c02' },
  ].filter(item => item.value > 0); // 只显示有数据的状态

  return (
    <Box>
      <Typography variant="h4" gutterBottom>
        仪表盘
      </Typography>
      <Typography variant="body1" color="text.secondary" paragraph>
        OpAMP Agent 管理平台概览
      </Typography>

      <Box sx={{ display: 'flex', gap: 3, flexWrap: 'wrap', mb: 3 }}>
        <Box sx={{ flex: '1 1 200px', minWidth: 200 }}>
          <StatCard
            title="总 Agents"
            value={total}
            icon={<AgentsIcon sx={{ color: 'white' }} />}
            color="#1976d2"
          />
        </Box>
        <Box sx={{ flex: '1 1 200px', minWidth: 200 }}>
          <StatCard
            title="在线 Agents"
            value={connectedAgents}
            icon={<ConnectedIcon sx={{ color: 'white' }} />}
            color="#2e7d32"
          />
        </Box>
        <Box sx={{ flex: '1 1 200px', minWidth: 200 }}>
          <StatCard
            title="离线 Agents"
            value={disconnectedAgents}
            icon={<DisconnectedIcon sx={{ color: 'white' }} />}
            color="#d32f2f"
          />
        </Box>
        <Box sx={{ flex: '1 1 200px', minWidth: 200 }}>
          <StatCard
            title="配置总数"
            value={configurations.length}
            icon={<ConfigIcon sx={{ color: 'white' }} />}
            color="#ed6c02"
          />
        </Box>
      </Box>

      {/* 图表区域 */}
      <Box sx={{ display: 'flex', gap: 3, flexWrap: 'wrap', mb: 3 }}>
        {/* Agent 状态分布饼图 */}
        <Box sx={{ flex: '1 1 400px' }}>
          <Paper sx={{ p: 3 }}>
            <Typography variant="h6" gutterBottom>
              Agent 状态分布
            </Typography>
            {statusData.length > 0 ? (
              <ResponsiveContainer width="100%" height={300}>
                <PieChart>
                  <Pie
                    data={statusData}
                    cx="50%"
                    cy="50%"
                    labelLine={false}
                    label={(entry: any) => `${entry.name} ${(entry.percent * 100).toFixed(0)}%`}
                    outerRadius={80}
                    fill="#8884d8"
                    dataKey="value"
                  >
                    {statusData.map((entry, index) => (
                      <Cell key={`cell-${index}`} fill={entry.color} />
                    ))}
                  </Pie>
                  <Tooltip />
                  <Legend />
                </PieChart>
              </ResponsiveContainer>
            ) : (
              <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: 300 }}>
                <Typography variant="body2" color="text.secondary">
                  暂无 Agent 数据
                </Typography>
              </Box>
            )}
          </Paper>
        </Box>

        {/* 占位：后续添加配置分发状态图 */}
        <Box sx={{ flex: '1 1 400px' }}>
          <Paper sx={{ p: 3 }}>
            <Typography variant="h6" gutterBottom>
              系统活动概览
            </Typography>
            <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2, mt: 2 }}>
              <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <Typography variant="body2">总 Agents</Typography>
                <Typography variant="h6" color="primary">{total}</Typography>
              </Box>
              <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <Typography variant="body2">在线率</Typography>
                <Typography variant="h6" color="success.main">
                  {total > 0 ? ((connectedAgents / total) * 100).toFixed(1) : 0}%
                </Typography>
              </Box>
              <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <Typography variant="body2">配置总数</Typography>
                <Typography variant="h6" color="warning.main">{configurations.length}</Typography>
              </Box>
              <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <Typography variant="body2">活跃配置</Typography>
                <Typography variant="h6" color="info.main">
                  {configurations.filter(c => c.selector && Object.keys(c.selector).length > 0).length}
                </Typography>
              </Box>
            </Box>
          </Paper>
        </Box>
      </Box>

      {/* 列表区域 */}
      <Box sx={{ display: 'flex', gap: 3, flexWrap: 'wrap' }}>
        <Box sx={{ flex: '1 1 400px' }}>
          <Paper sx={{ p: 3 }}>
            <Typography variant="h6" gutterBottom>
              最近连接的 Agents
            </Typography>
            {agents.length > 0 ? (
              agents.slice(0, 5).map((agent) => (
                <Box
                  key={agent.id}
                  sx={{
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                    py: 1,
                    borderBottom: '1px solid #eee',
                  }}
                >
                  <Box>
                    <Typography variant="body2" fontWeight="bold">
                      {agent.name || agent.id.substring(0, 8)}
                    </Typography>
                    <Typography variant="caption" color="text.secondary">
                      {agent.hostname}
                    </Typography>
                  </Box>
                  <Typography
                    variant="caption"
                    sx={{
                      px: 1,
                      py: 0.5,
                      borderRadius: 1,
                      backgroundColor:
                        agent.status === 'connected' ? '#e8f5e9' : '#ffebee',
                      color:
                        agent.status === 'connected' ? '#2e7d32' : '#d32f2f',
                    }}
                  >
                    {agent.status === 'connected' ? '在线' : '离线'}
                  </Typography>
                </Box>
              ))
            ) : (
              <Typography variant="body2" color="text.secondary">
                暂无 Agent 数据
              </Typography>
            )}
          </Paper>
        </Box>

        <Box sx={{ flex: '1 1 400px' }}>
          <Paper sx={{ p: 3 }}>
            <Typography variant="h6" gutterBottom>
              最近更新的配置
            </Typography>
            {configurations.length > 0 ? (
              configurations.slice(0, 5).map((config) => (
                <Box
                  key={config.id}
                  sx={{
                    py: 1,
                    borderBottom: '1px solid #eee',
                  }}
                >
                  <Typography variant="body2" fontWeight="bold">
                    {config.display_name}
                  </Typography>
                  <Typography variant="caption" color="text.secondary">
                    {config.name} • {config.content_type.toUpperCase()}
                  </Typography>
                </Box>
              ))
            ) : (
              <Typography variant="body2" color="text.secondary">
                暂无配置数据
              </Typography>
            )}
          </Paper>
        </Box>
      </Box>
    </Box>
  );
}
