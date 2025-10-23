import { useEffect, useState } from 'react';
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
  IconButton,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  Alert,
  Tooltip,
  Chip,
  TextField,
  Select,
  MenuItem,
  FormControl,
  InputLabel,
  CircularProgress,
} from '@mui/material';
import {
  Edit as EditIcon,
  Delete as DeleteIcon,
  Refresh as RefreshIcon,
  Add as AddIcon,
  History as HistoryIcon,
  Visibility as ViewIcon,
} from '@mui/icons-material';
import Editor from '@monaco-editor/react';
import { useConfigurationStore } from '@/stores/configurationStore';
import { CreateConfigurationRequest, UpdateConfigurationRequest, ConfigurationHistory } from '@/types/api';
import { format } from 'date-fns';
import axios from 'axios';

export default function ConfigurationListPage() {
  const {
    configurations,
    isLoading,
    error,
    fetchConfigurations,
    createConfiguration,
    updateConfiguration,
    deleteConfiguration,
    clearError,
  } = useConfigurationStore();

  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const [editDialogOpen, setEditDialogOpen] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [historyDialogOpen, setHistoryDialogOpen] = useState(false);
  const [selectedConfigName, setSelectedConfigName] = useState<string | null>(null);
  const [configHistories, setConfigHistories] = useState<ConfigurationHistory[]>([]);
  const [loadingHistory, setLoadingHistory] = useState(false);

  // 表单状态
  const [formData, setFormData] = useState({
    name: '',
    display_name: '',
    content_type: 'yaml' as 'yaml' | 'json',
    raw_config: '',
    selector: {} as Record<string, string>,
  });
  const [selectorKey, setSelectorKey] = useState('');
  const [selectorValue, setSelectorValue] = useState('');

  useEffect(() => {
    fetchConfigurations();
  }, [fetchConfigurations]);

  const handleCreateClick = () => {
    setFormData({
      name: '',
      display_name: '',
      content_type: 'yaml',
      raw_config: '',
      selector: {},
    });
    setCreateDialogOpen(true);
  };

  const handleEditClick = (config: any) => {
    setFormData({
      name: config.name,
      display_name: config.display_name,
      content_type: config.content_type,
      raw_config: config.raw_config,
      selector: config.selector || {},
    });
    setSelectedConfigName(config.name);
    setEditDialogOpen(true);
  };

  const handleDeleteClick = (name: string) => {
    setSelectedConfigName(name);
    setDeleteDialogOpen(true);
  };

  const handleCreateSubmit = async () => {
    try {
      const data: CreateConfigurationRequest = {
        name: formData.name,
        display_name: formData.display_name,
        content_type: formData.content_type,
        raw_config: formData.raw_config,
        selector: Object.keys(formData.selector).length > 0 ? formData.selector : undefined,
      };
      await createConfiguration(data);
      setCreateDialogOpen(false);
    } catch (err) {
      // 错误已在 store 中处理
    }
  };

  const handleEditSubmit = async () => {
    if (selectedConfigName) {
      try {
        const data: UpdateConfigurationRequest = {
          display_name: formData.display_name,
          raw_config: formData.raw_config,
          selector: Object.keys(formData.selector).length > 0 ? formData.selector : undefined,
        };
        await updateConfiguration(selectedConfigName, data);
        setEditDialogOpen(false);
      } catch (err) {
        // 错误已在 store 中处理
      }
    }
  };

  const handleDeleteConfirm = async () => {
    if (selectedConfigName) {
      try {
        await deleteConfiguration(selectedConfigName);
        setDeleteDialogOpen(false);
        setSelectedConfigName(null);
      } catch (err) {
        // 错误已在 store 中处理
      }
    }
  };

  const handleAddSelector = () => {
    if (selectorKey && selectorValue) {
      setFormData({
        ...formData,
        selector: { ...formData.selector, [selectorKey]: selectorValue },
      });
      setSelectorKey('');
      setSelectorValue('');
    }
  };

  const handleRemoveSelector = (key: string) => {
    const newSelector = { ...formData.selector };
    delete newSelector[key];
    setFormData({ ...formData, selector: newSelector });
  };

  const handleViewHistory = async (name: string) => {
    setSelectedConfigName(name);
    setHistoryDialogOpen(true);
    setLoadingHistory(true);
    try {
      const token = localStorage.getItem('token');
      const response = await axios.get(
        `http://localhost:8080/api/v1/configurations/${name}/history`,
        {
          headers: { Authorization: `Bearer ${token}` },
          params: { limit: 50, offset: 0 },
        }
      );
      setConfigHistories(response.data.histories || []);
    } catch (err) {
      console.error('Failed to fetch history:', err);
      setConfigHistories([]);
    } finally {
      setLoadingHistory(false);
    }
  };

  const ConfigDialog = ({
    open,
    onClose,
    onSubmit,
    title,
    isEdit,
  }: {
    open: boolean;
    onClose: () => void;
    onSubmit: () => void;
    title: string;
    isEdit: boolean;
  }) => (
    <Dialog open={open} onClose={onClose} maxWidth="md" fullWidth>
      <DialogTitle>{title}</DialogTitle>
      <DialogContent>
        <Box sx={{ mt: 2 }}>
          <TextField
            fullWidth
            label="配置名称 (Name)"
            value={formData.name}
            onChange={(e) => setFormData({ ...formData, name: e.target.value })}
            disabled={isEdit}
            margin="normal"
            required
            helperText="唯一标识,创建后不可修改"
          />
          <TextField
            fullWidth
            label="显示名称 (Display Name)"
            value={formData.display_name}
            onChange={(e) => setFormData({ ...formData, display_name: e.target.value })}
            margin="normal"
            required
          />
          <FormControl fullWidth margin="normal">
            <InputLabel>配置类型</InputLabel>
            <Select
              value={formData.content_type}
              label="配置类型"
              onChange={(e) =>
                setFormData({ ...formData, content_type: e.target.value as 'yaml' | 'json' })
              }
              disabled={isEdit}
            >
              <MenuItem value="yaml">YAML</MenuItem>
              <MenuItem value="json">JSON</MenuItem>
            </Select>
          </FormControl>

          <Typography variant="subtitle2" sx={{ mt: 2, mb: 1 }}>
            配置内容
          </Typography>
          <Box sx={{ border: '1px solid #ddd', borderRadius: 1, overflow: 'hidden' }}>
            <Editor
              height="300px"
              language={formData.content_type}
              value={formData.raw_config}
              onChange={(value) => setFormData({ ...formData, raw_config: value || '' })}
              options={{
                minimap: { enabled: false },
                fontSize: 13,
              }}
            />
          </Box>

          <Typography variant="subtitle2" sx={{ mt: 3, mb: 1 }}>
            标签选择器 (可选)
          </Typography>
          <Box sx={{ display: 'flex', gap: 1, mb: 1 }}>
            <TextField
              size="small"
              label="键"
              value={selectorKey}
              onChange={(e) => setSelectorKey(e.target.value)}
              sx={{ flex: 1 }}
            />
            <TextField
              size="small"
              label="值"
              value={selectorValue}
              onChange={(e) => setSelectorValue(e.target.value)}
              sx={{ flex: 1 }}
            />
            <Button variant="outlined" onClick={handleAddSelector}>
              添加
            </Button>
          </Box>
          <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 1 }}>
            {Object.entries(formData.selector).map(([key, value]) => (
              <Chip
                key={key}
                label={`${key}: ${value}`}
                onDelete={() => handleRemoveSelector(key)}
                variant="outlined"
              />
            ))}
          </Box>
        </Box>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>取消</Button>
        <Button onClick={onSubmit} variant="contained" disabled={isLoading}>
          {isEdit ? '更新' : '创建'}
        </Button>
      </DialogActions>
    </Dialog>
  );

  return (
    <Box>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Typography variant="h4">配置管理</Typography>
        <Box>
          <Tooltip title="刷新">
            <IconButton onClick={() => fetchConfigurations()} disabled={isLoading}>
              <RefreshIcon />
            </IconButton>
          </Tooltip>
          <Button
            variant="contained"
            startIcon={<AddIcon />}
            onClick={handleCreateClick}
            sx={{ ml: 1 }}
          >
            创建配置
          </Button>
        </Box>
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
              <TableCell>配置名称</TableCell>
              <TableCell>显示名称</TableCell>
              <TableCell>类型</TableCell>
              <TableCell>选择器</TableCell>
              <TableCell>更新时间</TableCell>
              <TableCell>操作</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {configurations.length === 0 && !isLoading ? (
              <TableRow>
                <TableCell colSpan={6} align="center">
                  <Typography variant="body2" color="text.secondary">
                    暂无配置数据
                  </Typography>
                </TableCell>
              </TableRow>
            ) : (
              configurations.map((config) => (
                <TableRow key={config.id} hover>
                  <TableCell>
                    <Typography variant="body2" sx={{ fontFamily: 'monospace' }}>
                      {config.name}
                    </Typography>
                  </TableCell>
                  <TableCell>{config.display_name}</TableCell>
                  <TableCell>
                    <Chip label={config.content_type.toUpperCase()} size="small" />
                  </TableCell>
                  <TableCell>
                    {config.selector && Object.keys(config.selector).length > 0 ? (
                      <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 0.5 }}>
                        {Object.entries(config.selector).map(([key, value]) => (
                          <Chip key={key} label={`${key}:${value}`} size="small" variant="outlined" />
                        ))}
                      </Box>
                    ) : (
                      '-'
                    )}
                  </TableCell>
                  <TableCell>
                    {format(new Date(config.updated_at), 'yyyy-MM-dd HH:mm:ss')}
                  </TableCell>
                  <TableCell>
                    <Tooltip title="编辑">
                      <IconButton size="small" onClick={() => handleEditClick(config)}>
                        <EditIcon fontSize="small" />
                      </IconButton>
                    </Tooltip>
                    <Tooltip title="查看历史">
                      <IconButton size="small" onClick={() => handleViewHistory(config.name)}>
                        <HistoryIcon fontSize="small" />
                      </IconButton>
                    </Tooltip>
                    <Tooltip title="删除">
                      <IconButton
                        size="small"
                        onClick={() => handleDeleteClick(config.name)}
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
      </TableContainer>

      {/* 创建对话框 */}
      <ConfigDialog
        open={createDialogOpen}
        onClose={() => setCreateDialogOpen(false)}
        onSubmit={handleCreateSubmit}
        title="创建配置"
        isEdit={false}
      />

      {/* 编辑对话框 */}
      <ConfigDialog
        open={editDialogOpen}
        onClose={() => setEditDialogOpen(false)}
        onSubmit={handleEditSubmit}
        title="编辑配置"
        isEdit={true}
      />

      {/* 删除确认对话框 */}
      <Dialog open={deleteDialogOpen} onClose={() => setDeleteDialogOpen(false)}>
        <DialogTitle>确认删除</DialogTitle>
        <DialogContent>
          <Typography>确定要删除这个配置吗?此操作不可撤销。</Typography>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDeleteDialogOpen(false)}>取消</Button>
          <Button onClick={handleDeleteConfirm} color="error" variant="contained">
            删除
          </Button>
        </DialogActions>
      </Dialog>

      {/* 历史版本对话框 */}
      <Dialog
        open={historyDialogOpen}
        onClose={() => setHistoryDialogOpen(false)}
        maxWidth="md"
        fullWidth
      >
        <DialogTitle>
          配置历史版本 - {selectedConfigName}
        </DialogTitle>
        <DialogContent>
          {loadingHistory ? (
            <Box sx={{ display: 'flex', justifyContent: 'center', py: 4 }}>
              <CircularProgress />
            </Box>
          ) : configHistories.length === 0 ? (
            <Typography variant="body2" color="text.secondary" align="center" sx={{ py: 4 }}>
              暂无历史版本
            </Typography>
          ) : (
            <Table size="small">
              <TableHead>
                <TableRow>
                  <TableCell>版本</TableCell>
                  <TableCell>配置 Hash</TableCell>
                  <TableCell>创建时间</TableCell>
                  <TableCell>操作</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {configHistories.map((history) => (
                  <TableRow key={history.id} hover>
                    <TableCell>
                      <Chip label={`v${history.version}`} size="small" color="primary" />
                    </TableCell>
                    <TableCell>
                      <Typography variant="body2" sx={{ fontFamily: 'monospace', fontSize: '0.85rem' }}>
                        {history.config_hash.substring(0, 12)}...
                      </Typography>
                    </TableCell>
                    <TableCell>
                      {format(new Date(history.created_at), 'yyyy-MM-dd HH:mm:ss')}
                    </TableCell>
                    <TableCell>
                      <Tooltip title="查看配置">
                        <IconButton
                          size="small"
                          onClick={() => {
                            setFormData({
                              name: history.configuration_name,
                              display_name: '',
                              content_type: history.content_type,
                              raw_config: history.raw_config,
                              selector: history.selector || {},
                            });
                            setHistoryDialogOpen(false);
                            setEditDialogOpen(true);
                          }}
                        >
                          <ViewIcon fontSize="small" />
                        </IconButton>
                      </Tooltip>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setHistoryDialogOpen(false)}>关闭</Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}
