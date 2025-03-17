import React from 'react';
import ProLayout, {PageContainer} from '@ant-design/pro-layout';
import zhCN from 'antd/lib/locale/zh_CN';
import en from 'antd/lib/locale/en_US';
import {getLang, getLocale} from "../locale/locale";
import {Button, ConfigProvider, notification} from "antd";
import version from "../config/version.json";
import ReactMarkdown from "react-markdown";
import i18n from "i18next";
import axios from "axios";
import './wrapper.css';

function wrapper(props) {
	return (
		<ProLayout
			loading={false}
			title='FeArKit'
			logo={null}
			layout='top'
			navTheme='light'
			collapsed={true}
			fixedHeader={true}
			contentWidth='fluid'
			collapsedButtonRender={Title}
		>
			<PageContainer>
				<ConfigProvider locale={getLang()==='zh-CN'?zhCN:en}>
					{props.children}
				</ConfigProvider>
			</PageContainer>
		</ProLayout>
	);
}

function Title() {
	return (
		<div
			style={{
				userSelect: 'none',
				fontWeight: 500
			}}
		>
			FeArKit
		</div>
	)
}



export default wrapper;