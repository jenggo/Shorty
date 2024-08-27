<tbody id="tbody">
  {{range $val := .}}
  <tr>
    <td>{{$val.Shorty}}</td>
    <td>{{$val.File}}</td>
    <td>{{$val.Url}}</td>
    <td>{{$val.Expired}}</td>
  </tr>
  {{end}}
</tbody>
