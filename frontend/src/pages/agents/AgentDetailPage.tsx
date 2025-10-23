import { useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
  Box,
  Paper,
  Typography,
  Chip,
  Button,
  Card,
  CardContent,
  Divider,
  Alert,
  CircularProgress,
} from '@mui/material';
import { ArrowBack as BackIcon } from '@mui/icons-material';
import { useAgentStore } from '@/stores/agentStore';
import { format } from 'date-fns';

export default function AgentDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { selectedAgent, isLoading, error, fetchAgent, clearError } = useAgentStore();

  useEffect(() => {
    if (id) {
      fetchAgent(id);
    }
  }, [id, fetchAgent]);

  if (isLoading) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: 400 }}>
        <CircularProgress />
      </Box>
    );
  }

  if (error) {
    return (
      <Box>
        <Alert severity="error" onClose={clearError}>
          {error}
        </Alert>
        <Button startIcon={<BackIcon />} onClick={() => navigate('/agents')} sx={{ mt: 2 }}>
          返回列表
        </Button>
      </Box>
    );
  }

  if (!selectedAgent) {
    return null;
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'connected':
        return 'success';
      case 'disconnected':
        return 'default';
      case 'configuring':
        return 'warning';
      case 'error':
        return 'error';
      default:
        return 'default';
    }
  };

  const getStatusText = (status: string) => {
    switch (status) {
      case 'connected':
        return '在线';
      case 'disconnected':
        return '离线';
      case 'configuring':
        return '配置中';
      case 'error':
        return '错误';
      default:
        return status;
    }
  };

  return (
    <Box>
      <Button startIcon={<BackIcon />} onClick={() => navigate('/agents')} sx={{ mb: 2 }}>
        返回列表
      </Button>

      <Paper sx={{ p: 3 }}>
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
          <Typography variant="h4">Agent 详情</Typography>
          <Chip
            label={getStatusText(selectedAgent.status)}
            color={getStatusColor(selectedAgent.status)}
          />
        </Box>

        <Box sx={{ display: 'flex', gap: 3, flexWrap: 'wrap' }}>
          <Box sx={{ flex: '1 1 400px' }}>
            <Card variant="outlined">
              <CardContent>
                <Typography variant="h6" gutterBottom>
                  基本信息
                </Typography>
                <Divider sx={{ mb: 2 }} />
                <Box sx={{ mb: 2 }}>
                  <Typography variant="caption" color="text.secondary">
                    Agent ID
                  </Typography>
                  <Typography
                    variant="body2"
                    sx={{ fontFamily: 'monospace', wordBreak: 'break-all' }}
                  >
                    {selectedAgent.id}
                  </Typography>
                </Box>
                <Box sx={{ mb: 2 }}>
                  <Typography variant="caption" color="text.secondary">
                    名称
                  </Typography>
                  <Typography variant="body2">{selectedAgent.name || '-'}</Typography>
                </Box>
                <Box sx={{ mb: 2 }}>
                  <Typography variant="caption" color="text.secondary">
                    版本
                  </Typography>
                  <Typography variant="body2">{selectedAgent.version}</Typography>
                </Box>
                <Box sx={{ mb: 2 }}>
                  <Typography variant="caption" color="text.secondary">
                    主机名
                  </Typography>
                  <Typography variant="body2">{selectedAgent.hostname}</Typography>
                </Box>
              </CardContent>
            </Card>
          </Box>

          <Box sx={{ flex: '1 1 400px' }}>
            <Card variant="outlined">
              <CardContent>
                <Typography variant="h6" gutterBottom>
                  系统信息
                </Typography>
                <Divider sx={{ mb: 2 }} />
                <Box sx={{ mb: 2 }}>
                  <Typography variant="caption" color="text.secondary">
                    操作系统
                  </Typography>
                  <Typography variant="body2">{selectedAgent.os_type}</Typography>
                </Box>
                <Box sx={{ mb: 2 }}>
                  <Typography variant="caption" color="text.secondary">
                    架构
                  </Typography>
                  <Typography variant="body2">{selectedAgent.architecture}</Typography>
                </Box>
                <Box sx={{ mb: 2 }}>
                  <Typography variant="caption" color="text.secondary">
                    最后连接时间
                  </Typography>
                  <Typography variant="body2">
                    {selectedAgent.last_seen
                      ? format(new Date(selectedAgent.last_seen), 'yyyy-MM-dd HH:mm:ss')
                      : '-'}
                  </Typography>
                </Box>
                <Box sx={{ mb: 2 }}>
                  <Typography variant="caption" color="text.secondary">
                    创建时间
                  </Typography>
                  <Typography variant="body2">
                    {format(new Date(selectedAgent.created_at), 'yyyy-MM-dd HH:mm:ss')}
                  </Typography>
                </Box>
              </CardContent>
            </Card>
          </Box>

          {selectedAgent.labels && Object.keys(selectedAgent.labels).length > 0 && (
            <Box sx={{ flex: '1 1 100%' }}>
              <Card variant="outlined">
                <CardContent>
                  <Typography variant="h6" gutterBottom>
                    标签
                  </Typography>
                  <Divider sx={{ mb: 2 }} />
                  <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 1 }}>
                    {Object.entries(selectedAgent.labels).map(([key, value]) => (
                      <Chip key={key} label={`${key}: ${value}`} variant="outlined" />
                    ))}
                  </Box>
                </CardContent>
              </Card>
            </Box>
          )}

          {selectedAgent.effective_config && (
            <Box sx={{ flex: '1 1 100%' }}>
              <Card variant="outlined">
                <CardContent>
                  <Typography variant="h6" gutterBottom>
                    当前配置
                  </Typography>
                  <Divider sx={{ mb: 2 }} />
                  <Box
                    component="pre"
                    sx={{
                      backgroundColor: '#f5f5f5',
                      p: 2,
                      borderRadius: 1,
                      overflow: 'auto',
                      maxHeight: 400,
                      fontFamily: 'monospace',
                      fontSize: '0.875rem',
                    }}
                  >
                    {selectedAgent.effective_config}
                  </Box>
                </CardContent>
              </Card>
            </Box>
          )}
        </Box>
      </Paper>
    </Box>
  );
}
