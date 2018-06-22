import React from 'react'
import { Highcharts } from 'react-highcharts'
import HighchartsMore from 'highcharts-more'
import ReactHighstock from 'react-highcharts/ReactHighstock.src'

// controls data flow to the chart component
export default function ChartView(props) {
    const { config } = props
    return (
        <div>
            <ReactHighstock config={config}/>
        </div>
    )
}