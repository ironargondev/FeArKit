import React from 'react';
import {ModalForm, ProFormText} from '@ant-design/pro-form';
import {request} from "../../utils/utils";
import i18n from "../../locale/locale";
import {message} from "antd";

function Executable(props) {
	async function onFinish(form) {
		form.device = props.device.id;
		let basePath = location.origin + location.pathname + 'api/device/';
		request(basePath + 'executable', form).then(res => {
			if (res.data.code === 0) {
				message.success(i18n.t('EXECUTE.EXECUTION_SUCCESS'));
			}
		});
	}

	return (
		<ModalForm
			modalProps={{
				destroyOnClose: true,
				maskClosable: true,
			}}
			title={i18n.t('EXECUTE.TITLE')}
			width={380}
			onFinish={onFinish}
			onVisibleChange={open => {
				if (!open) props.onCancel();
			}}
			submitter={{
				render: (_, elems) => elems.pop()
			}}
			{...props}
		>
			<ProFormText
				width="md"
				name="url"
				label={'URL'}
				rules={[{
					required: true
				}]}
			/>
			<ProFormText
				width="md"
				name="path"
				label={"path"}
				description={"path to save the file. Defaults to temporary directory."}
			/>
		</ModalForm>
	)
}

export default Executable;