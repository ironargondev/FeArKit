import React, { useState } from "react";
import { Button } from "antd";
import { UploadOutlined } from "@ant-design/icons";
import { ModalForm } from '@ant-design/pro-form';
import FileUploader from "./uploader";


let uploaderEndoint = '/api/device/loadelf?'

function Loadelf(props) {
	const [uploading, setUploading] = useState(false);
	const [path, setPath] = useState(false);

	function uploadFile() {
		document.getElementById('file-uploader').click();
	}
	function onFileChange(e) {
		let file = e.target.files[0];
		if (file === undefined) return;
		e.target.value = null;
		setUploading(file);
	}
	function onUploadSuccess() {
		setUploading(false);
	}
	function onUploadCancel() {
		setUploading(false);
	}

	return (
		<ModalForm
			modalProps={{
				destroyOnClose: true,
				maskClosable: true,
			}}
			onVisibleChange={open => {
				if (!open) props.onCancel();
			}}
			title={"Load Elf"}
			width={380}
			submitter={false}

			{...props}
		>
			<Button
				icon={<UploadOutlined />}
				onClick={uploadFile}
			/>
			<input
				id='file-uploader'
				type='file'
				onChange={onFileChange}
				style={{ display: 'none' }}
			/>
			<FileUploader
				open={uploading}
				uploaderEndpoint={uploaderEndoint}
				file={uploading}
				device={props.device}
				onSuccess={onUploadSuccess}
				onCancel={onUploadCancel}
			/>
		</ModalForm>
	)
}


export default Loadelf;