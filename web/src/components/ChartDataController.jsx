import React, { Component } from 'react'
import ChartController from './ChartController'

// pulls chart data from the server (or demo)
export default class ChartDataController extends Component {
    state = {
        blocks: []
    }

    componentDidMount() {
        
    }

    render() {
        return (
        <div>
            <ChartController />
        </div>
        );
    }

    demoData() {
        const block1 = {
            action: "issue",
            account: "xtb:LS0tLS1CRUdJTiBSU0EgUFVCTElDIEtFWS0tLS0tCglNSUdmTUEwR0NTcUdTSWIzRFFFQkFRVUFBNEdOQURDQmlRS0JnUURWbGVZUStNT0doSFZ2a216Q2tKcmpJNUNMCgk0Tk1Id05SbDdTUm5FbEZJMituV2pZTUV3U09scDVwVGNIQnpqUmhKT3gxU2JMdGlLUktGZzFROXdVZXZOZVdTCglQTWpCMWwrTFdtVVRScU5UY0FQUWMwVmRldW1qcXMxUCtlSEVSZms5TXdxTnNyUHl0dkd3dk5RSjA1UGtnTFNrCglYdTU4a3I1aVh4TUFCSXVrYlFJREFRQUIKCS0tLS0tRU5EIFJTQSBQVUJMSUMgS0VZLS0tLS0KCQ",
            token: "xtb:LS0tLS1CRUdJTiBSU0EgUFVCTElDIEtFWS0tLS0tCglNSUdmTUEwR0NTcUdTSWIzRFFFQkFRVUFBNEdOQURDQmlRS0JnUURWbGVZUStNT0doSFZ2a216Q2tKcmpJNUNMCgk0Tk1Id05SbDdTUm5FbEZJMituV2pZTUV3U09scDVwVGNIQnpqUmhKT3gxU2JMdGlLUktGZzFROXdVZXZOZVdTCglQTWpCMWwrTFdtVVRScU5UY0FQUWMwVmRldW1qcXMxUCtlSEVSZms5TXdxTnNyUHl0dkd3dk5RSjA1UGtnTFNrCglYdTU4a3I1aVh4TUFCSXVrYlFJREFRQUIKCS0tLS0tRU5EIFJTQSBQVUJMSUMgS0VZLS0tLS0KCQ",
            previous: "placholder_previou1s",
            representative: "placholder_representative1",
            balance: 100,
            link: "placholder_link1",
            signature: "placholder_signature1",
            hash: "hash_1"
        }
        const block2 = {
            action: "send",
            account: "xtb:LS0tLS1CRUdJTiBSU0EgUFVCTElDIEtFWS0tLS0tCglNSUdmTUEwR0NTcUdTSWIzRFFFQkFRVUFBNEdOQURDQmlRS0JnUURWbGVZUStNT0doSFZ2a216Q2tKcmpJNUNMCgk0Tk1Id05SbDdTUm5FbEZJMituV2pZTUV3U09scDVwVGNIQnpqUmhKT3gxU2JMdGlLUktGZzFROXdVZXZOZVdTCglQTWpCMWwrTFdtVVRScU5UY0FQUWMwVmRldW1qcXMxUCtlSEVSZms5TXdxTnNyUHl0dkd3dk5RSjA1UGtnTFNrCglYdTU4a3I1aVh4TUFCSXVrYlFJREFRQUIKCS0tLS0tRU5EIFJTQSBQVUJMSUMgS0VZLS0tLS0KCQ",
            token: "xtb:LS0tLS1CRUdJTiBSU0EgUFVCTElDIEtFWS0tLS0tCglNSUdmTUEwR0NTcUdTSWIzRFFFQkFRVUFBNEdOQURDQmlRS0JnUURWbGVZUStNT0doSFZ2a216Q2tKcmpJNUNMCgk0Tk1Id05SbDdTUm5FbEZJMituV2pZTUV3U09scDVwVGNIQnpqUmhKT3gxU2JMdGlLUktGZzFROXdVZXZOZVdTCglQTWpCMWwrTFdtVVRScU5UY0FQUWMwVmRldW1qcXMxUCtlSEVSZms5TXdxTnNyUHl0dkd3dk5RSjA1UGtnTFNrCglYdTU4a3I1aVh4TUFCSXVrYlFJREFRQUIKCS0tLS0tRU5EIFJTQSBQVUJMSUMgS0VZLS0tLS0KCQ",
            previous: "HYC4A7ZVRDZW4ZF5UZH2JJKJ7BK6HZYNA6Y2TIGDTZSZAF6OIIRA",
            representative: "placholder_previous2",
            balance: 100,
            link: "xtb:testreceiver",
            signature: "placholder_signature2",
            hash: "hash_2"
        }
        const demoBlocks = [block1, block2]
        return demoBlocks
    }
}