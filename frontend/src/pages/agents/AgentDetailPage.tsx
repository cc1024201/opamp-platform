import { useEffect, useState } from 'react';
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
  IconButton,
  Tooltip,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
} from '@mui/material';
import {
  ArrowBack as BackIcon,
  Refresh as RefreshIcon,
  Delete as DeleteIcon,
  ContentCopy as CopyIcon,
} from '@mui/icons-material';
import { useAgentStore } from '@/stores/agentStore';
import { format } from 'date-fns';

export default function AgentDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { selectedAgent, isLoading, error, fetchAgent, deleteAgent, clearError } = useAgentStore();
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [copySuccess, setCopySuccess] = useState(false);

  useEffect(() => {
    if (id) {
      fetchAgent(id);
    }
  }, [id, fetchAgent]);

  const handleRefresh = () => {
    if (id) {
      fetchAgent(id);
    }
  };

  const handleDelete = async () => {
    if (id) {
      try {
        await deleteAgent(id);
        navigate('/agents');
      } catch (err) {
        // 错误已在 store 中处理
      }
    }
  };

  const handleCopyId = () => {
    if (selectedAgent) {
      navigator.clipboard.writeText(selectedAgent.id);
      setCopySuccess(true);
      setTimeout(() => setCopySuccess(false), 2000);
    }
  };

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
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
        <Button startIcon={<BackIcon />} onClick={() => navigate('/agents')}>
          返回列表
        </Button>
        <Box sx={{ display: 'flex', gap: 1 }}>
          <Tooltip title="刷新">
            <IconButton onClick={handleRefresh} disabled={isLoading}>
              <RefreshIcon />
            </IconButton>
          </Tooltip>
          <Tooltip title="删除 Agent">
            <IconButton onClick={() => setDeleteDialogOpen(true)} color="error">
              <DeleteIcon />
            </IconButton>
          </Tooltip>
        </Box>
      </Box>

      {copySuccess && (
        <Alert severity="success" sx={{ mb: 2 }}>
          Agent ID 已复制到剪贴板
        </Alert>
      )}

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
                  <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                    <Typography variant="caption" color="text.secondary">
                      Agent ID
                    </Typography>
                    <Tooltip title={copySuccess ? '已复制!' : '复制 ID'}>
                      <IconButton size="small" onClick={handleCopyId}>
                        <CopyIcon fontSize="small" />
                      </IconButton>
                    </Tooltip>
                  </Box>
                  <Typography
                    variant="body2"
                    sx={{ fontFamily: 'monospace', wordBreak: 'break-all', mt: 0.5 }}
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

      {/* 删除确认对话框 */}
      <Dialog open={deleteDialogOpen} onClose={() => setDeleteDialogOpen(false)}>
        <DialogTitle>确认删除</DialogTitle>
        <DialogContent>
          <Typography>
            确定要删除 Agent <strong>{selectedAgent.name || selectedAgent.id.substring(0, 8)}</strong> 吗?
          </Typography>
          <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
            此操作不可撤销,所有相关数据将被永久删除。
          </Typography>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDeleteDialogOpen(false)}>取消</Button>
          <Button onClick={handleDelete} color="error" variant="contained">
            删除
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}
