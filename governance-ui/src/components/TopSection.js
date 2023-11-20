import { Box, Typography, Avatar, Grid } from "@mui/material";
import logo from "../assessts/vitwit-logo.jpg";

function TopSection({ from, to }) {
  return (
      <Box
        sx={{
          background: "#efefef"
        }}
      >
        <Grid container alignItems="center" justifyContent="center" spacing={2}>
          <Grid
            item
            style={{
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
            }}
          >
            <Avatar
              alt="Witval"
              src={logo}
              style={{ width: "65px", height: "65px" }}
            />
          </Grid>
          <Grid item>
            <Typography variant="h4" fontWeight={600}>
              |
            </Typography>
          </Grid>
          <Grid item>
            <Typography variant="h4" fontWeight={600}>
              Witval Governance Report
            </Typography>
          </Grid>
          <Grid item>
            <Typography variant="h4">|</Typography>
          </Grid>
          <Grid item>
            <Typography variant="h4" fontWeight={600}>
              {from ? from : ""}&nbsp;-&nbsp;
              {to ? to : ""}
            </Typography>
          </Grid>
        </Grid>
      </Box>
  );
}

export default TopSection;
