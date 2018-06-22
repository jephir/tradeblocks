import React, { Component } from 'react'
import ChartView from './ChartView'
import DemoData from '../DemoData'

// controls data flow to the chart component
export default class ChartController extends Component {
    state = {
        config: {}
    }

    componentDidMount() {
        const data = JSON.parse(JSON.stringify(DemoData))
        console.log(data)
        var ohlc = [],
        volume = [],
        dataLength = data.length,
        // set the allowed units for data grouping
        groupingUnits = [[
            'week',                         // unit name
            [1]                             // allowed multiples
        ], [
            'month',
            [1, 2, 3, 4, 6]
        ]],

        i = 0;

        for (i; i < dataLength; i += 1) {
            ohlc.push([
                data[i][0], // the date
                data[i][1], // open
                data[i][2], // high
                data[i][3], // low
                data[i][4] // close
            ]);

            volume.push([
                data[i][0], // the date
                data[i][5] // the volume
            ]);
        }
        var config = {
            rangeSelector: {
                selected: 1
            },
    
            title: {
                text: 'XTB PRice'
            },
    
            yAxis: [{
                labels: {
                    align: 'right',
                    x: -3
                },
                title: {
                    text: 'OHLC'
                },
                height: '60%',
                lineWidth: 2,
                resize: {
                    enabled: true
                }
            }, {
                labels: {
                    align: 'right',
                    x: -3
                },
                title: {
                    text: 'Volume'
                },
                top: '65%',
                height: '35%',
                offset: 0,
                lineWidth: 2
            }],
    
            tooltip: {
                split: true
            },
    
            series: [{
                type: 'candlestick',
                name: 'XTB',
                data: ohlc,
                dataGrouping: {
                    units: groupingUnits
                }
            }, {
                type: 'column',
                name: 'Volume',
                data: volume,
                yAxis: 1,
                dataGrouping: {
                    units: groupingUnits
                }
            }]
        }
        this.setState({ config })
    }

    render() {
        const { config } = this.state
        return <ChartView config={config} />
    }
}