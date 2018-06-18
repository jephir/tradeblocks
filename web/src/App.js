import React, { Component } from 'react';
import './App.css';
import Tabs from '@material-ui/core/Tabs'
import Tab from '@material-ui/core/Tab'
import BlockDataController from './components/BlockDataController.jsx'
import ChartDataController from './components/ChartDataController.jsx'

class App extends Component {
  state = {
    value: 1,
  }

  handleChange = (event, value) => {
    this.setState({ value });
  }

  render() {
    const { value } = this.state
    return (
      <div>
        <Tabs value={value} onChange={this.handleChange}
              indicatorColor="primary">
          <Tab label={"Blocks"} />
          <Tab label={"Charts"} />
        </Tabs>
        {value === 0 && <BlockDataController />}
        {value === 1 && <ChartDataController />}
      </div>
    )
  }
}

export default App;
