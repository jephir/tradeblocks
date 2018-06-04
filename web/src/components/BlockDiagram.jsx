import React from 'react'
import PropTypes from 'prop-types'
import { withStyles } from '@material-ui/core/styles'
import Typography from '@material-ui/core/Typography'
import ReactJson from 'react-json-view'

const styles = {

}

function BlockDiagrm(props) {
    const {
        block
    } = props
    return <div>{block ? <ReactJson src={block} /> : "help"}</div>
}

BlockDiagrm.propTypes = {
    classes: PropTypes.object.isRequired,
}
  
export default withStyles(styles)(BlockDiagrm)