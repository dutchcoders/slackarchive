import axios from 'axios';

const apiUrl = window.location.protocol + '//' + window.location.host + '/v1/';

export default {
  getTeams(teamDomain){
    let params = {};
    if (teamDomain)
      params.domain = teamDomain;
    return axios.get(apiUrl + 'team', {params: params})
  },
  getChannels(teamId){
    return axios.get(apiUrl + 'channels', {params: {team_id: teamId}})
  },
  getMessages (team_id, channel_id, size, offset, search, sort = 'desc', tsTo, tsFrom) {
    let params = {size: size, team: team_id};
    if (channel_id)
      params.channel = channel_id
    if (offset !== undefined)
      params.offset = offset
    if (search) {
      params.q = search.query
      if (size > 0) {
        params.aggs = 1
      }
    }
    if (sort === 'asc')
      params.sort = sort
    if (tsTo)
      params.to = tsTo/1000000

    return axios.get(apiUrl + 'messages', {params: params})
  }
}
