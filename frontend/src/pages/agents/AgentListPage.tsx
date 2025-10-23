import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Box,
  Paper,
  Typography,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  TablePagination,
  Chip,
  IconButton,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  Alert,
  Tooltip,
} from '@mui/material';
import {
  Visibility as ViewIcon,
  Delete as DeleteIcon,
  Refresh as RefreshIcon,
} from '@mui/icons-material';
import { useAgentStore } from '@/stores/agentStore';
import { format } from 'date-fns';

export default function AgentListPage() {
  const navigate = useNavigate();
  const { agents, total, page, pageSize, isLoading, error, fetchAgents, deleteAgent, clearError } =
    useAgentStore();
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [selectedAgentId, setSelectedAgentId] = useState<string | null>(null);

  useEffect(() => {
    fetchAgents(page, pageSize);
  }, [fetchAgents, page, pageSize]);

  const handleChangePage = (_event: unknown, newPage: number) => {
    fetchAgents(newPage + 1, pageSize);
  };

  const handleChangeRowsPerPage = (event: React.ChangeEvent<HTMLInputElement>) => {
    fetchAgents(1, parseInt(event.target.value, 10));
  };

  const handleDeleteClick = (id: string) => {
    setSelectedAgentId(id);
    setDeleteDialogOpen(true);
  };

  const handleDeleteConfirm = async () => {
    if (selectedAgentId) {
      try {
        await deleteAgent(selectedAgentId);
        setDeleteDialogOpen(false);
        setSelectedAgentId(null);
      } catch (err) {
        // 错误已在 store 中处理
      }
    }
  };

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
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Typography variant="h4">Agent 管理</Typography>
        <Tooltip title="刷新">
          <IconButton onClick={() => fetchAgents(page, pageSize)} disabled={isLoading}>
            <RefreshIcon />
          </IconButton>
        </Tooltip>
      </Box>

      {error && (
        <Alert severity="error" onClose={clearError} sx={{ mb: 2 }}>
          {error}
        </Alert>
      )}

      <TableContainer component={Paper}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>Agent ID</TableCell>
              <TableCell>名称</TableCell>
              <TableCell>主机名</TableCell>
              <TableCell>版本</TableCell>
              <TableCell>操作系统</TableCell>
              <TableCell>架构</TableCell>
              <TableCell>状态</TableCell>
              <TableCell>最后连接</TableCell>
              <TableCell>操作</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {agents.length === 0 && !isLoading ? (
              <TableRow>
                <TableCell colSpan={9} align="center">
                  <Typography variant="body2" color="text.secondary">
                    暂无 Agent 数据
                  </Typography>
                </TableCell>
              </TableRow>
            ) : (
              agents.map((agent) => (
                <TableRow key={agent.id} hover>
                  <TableCell>
                    <Typography variant="body2" sx={{ fontFamily: 'monospace', fontSize: '0.85rem' }}>
                      {agent.id.substring(0, 8)}...
                    </Typography>
                  </TableCell>
                  <TableCell>{agent.name || '-'}</TableCell>
                  <TableCell>{agent.hostname}</TableCell>
                  <TableCell>{agent.version}</TableCell>
                  <TableCell>{agent.os_type}</TableCell>
                  <TableCell>{agent.architecture}</TableCell>
                  <TableCell>
                    <Chip
                      label={getStatusText(agent.status)}
                      color={getStatusColor(agent.status)}
                      size="small"
                    />
                  </TableCell>
                  <TableCell>
                    {agent.last_seen
                      ? format(new Date(agent.last_seen), 'yyyy-MM-dd HH:mm:ss')
                      : '-'}
                  </TableCell>
                  <TableCell>
                    <Tooltip title="查看详情">
                      <IconButton
                        size="small"
                        onClick={() => navigate(`/agents/${agent.id}`)}
                      >
                        <ViewIcon fontSize="small" />
                      </IconButton>
                    </Tooltip>
                    <Tooltip title="删除">
                      <IconButton
                        size="small"
                        onClick={() => handleDeleteClick(agent.id)}
                        color="error"
                      >
                        <DeleteIcon fontSize="small" />
                      </IconButton>
                    </Tooltip>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
        <TablePagination
          rowsPerPageOptions={[5, 10, 25, 50]}
          component="div"
          count={total}
          rowsPerPage={pageSize}
          page={page - 1}
          onPageChange={handleChangePage}
          onRowsPerPageChange={handleChangeRowsPerPage}
          labelRowsPerPage="每页显示:"
          labelDisplayedRows={({ from, to, count }) => `${from}-${to} 共 ${count} 条`}
        />
      </TableContainer>

      {/* 删除确认对话框 */}
      <Dialog open={deleteDialogOpen} onClose={() => setDeleteDialogOpen(false)}>
        <DialogTitle>确认删除</DialogTitle>
        <DialogContent>
          <Typography>确定要删除这个 Agent 吗?此操作不可撤销。</Typography>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDeleteDialogOpen(false)}>取消</Button>
          <Button onClick={handleDeleteConfirm} color="error" variant="contained">
            删除
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}
