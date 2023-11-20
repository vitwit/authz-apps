import {
  Card,
  CardContent,
  Typography,
  Avatar,
  Box,
} from "@mui/material";
import Text from "./Text";
import { NAME_TO_ASSET } from "./Data";

function CustomCard({ networkName, data }) {
  return (
    <Card
      variant="outlined"
      sx={{
        p: 1,
        borderRadius: 2,
      }}
    >
      <CardContent>
        <Box
          sx={{
            display: "flex",
            alignContent: "center",
            margin: "auto",
          }}
        >
          <Avatar
            alt={networkName}
            src={NAME_TO_ASSET[networkName?.toLowerCase()]}
          />
          <Typography
            variant="h5"
            component="div"
            fontWeight={600}
            gutterBottom
            sx={{
              mt: 0.5,
            }}
          >
            &nbsp;&nbsp;{networkName.charAt(0).toUpperCase() + networkName.slice(1)}
          </Typography>
        </Box>
        {}
        {data.map((item, index) => (
          <Box mt={1} key={index}>
            <Text
              proposalId={item?.proposalID}
              title={item?.title}
              voteOption={item?.vote_option}
            />
          </Box>
        ))}
      </CardContent>
    </Card>
  );
}

export default CustomCard;
