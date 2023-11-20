import TopSection from "./TopSection";
import CardList from "./CardList";
import { Box } from "@mui/system";

function MainPage({ from, to }) {
  const formatDate = (date) => {
    const day = date.date();
    const month = date.format("MMM");
    return `${day}/${month}`;
  };

  return (
    <div style={{ backgroundColor: "#efefef" }}>
      <TopSection from={formatDate(from)} to={formatDate(to)} />
      <Box
        sx={{
          mt: 2,
          mr: 4,
          ml: 4,
          mb: 2,
        }}
      >
        <CardList from={from} to={to} />
      </Box>
    </div>
  );
}

export default MainPage;
