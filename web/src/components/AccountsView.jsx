import React from 'react'
import PropTypes from 'prop-types'
import { withStyles } from '@material-ui/core/styles'
import Card from '@material-ui/core/Card'
import CardContent from '@material-ui/core/CardContent'
import Typography from '@material-ui/core/Typography'
import Grid from '@material-ui/core/Grid'

import BlocksController from './BlocksController'
import BlockDiagram from './BlockDiagram'

const styles = {
    accountHeader: {
        fontSize: 16
    },
    cardDiv: {
        display: "block",
        marginLeft: "px",
        marginRight: "auto",
        textAlign: "center",
        marginTop: "15px",
        wordBreak: "all",
    },
    grid: {
        flexGrow: 1
    }
}

function AccountsView(props) {
    const { 
        accounts,
        allBlocks,
        activeAccount,
        classes,
        handleClick,
    } = props
    if (accounts === undefined) {
        return <div>Loading...</div>
    }

    const accountsRender = Object.keys(accounts).map((key) => {
        const blocks = accounts[key]
        const shortAccount = key.substring(0,12) + "..."
        return (
            <div key={key} className={classes.cardDiv} onClick={() => handleClick(key)}>
                <Card className={classes.accountCard}>
                    <CardContent>
                        <Typography className={classes.accountHeader} nowrap={"true"}>
                            Account: {shortAccount}
                        </Typography>
                        <Typography component="p">
                            Number of Blocks: {blocks.length}
                        </Typography>
                    </CardContent>
                </Card>
            </div>
        )
    })
    
    const accountBlocks = accounts[activeAccount]
    // deep copy to trigger updates
    var filteredBlocks = []
    if (accountBlocks !== undefined) {
        filteredBlocks = accountBlocks.map((block) =>{
            return Object.assign({}, block)
        })
    }
    return (
        <div className={classes.grid}>
            <Grid container spacing={24}>
                <Grid item xs={2}>
                    {accountsRender}
                </Grid>
                <Grid item xs={2}>
                    <BlocksController filteredBlocks={filteredBlocks} allBlocks={allBlocks}/>
                </Grid>
                <Grid item xs={8} className={classes.diagram}>
                    <BlockDiagram allBlocks={allBlocks} accounts={accounts}/>
                </Grid>
            </Grid>
        </div>
    )
}

AccountsView.propTypes = {
    classes: PropTypes.object.isRequired,
  };
  
  export default withStyles(styles)(AccountsView);